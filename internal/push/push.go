package push

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sisyphus/internal/ado"
	"sisyphus/internal/payload"
)

type RunConfig struct {
	Issue       string
	BuildDef    string
	StartBuildID *int
	RepoPath    string
	LLM         string
	SleepSeconds int
	LogMaxBytes int
	ADOOrg      string
	ADOProject  string
	ADOBaseURL  string
	PAT         string
}

func runCmd(cwd string, cmd []string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}
	command := exec.Command(cmd[0], cmd[1:]...)
	if cwd != "" {
		command.Dir = cwd
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("command failed: %s\nstdout:\n%s\nstderr:\n%s", strings.Join(cmd, " "), stdout.String(), stderr.String())
	}
	return nil
}

func gitStatus(repoPath string) (string, error) {
	command := exec.Command("git", "status", "--porcelain")
	command.Dir = repoPath
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("git status failed: %s", strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func ensureClean(repoPath string) error {
	status, err := gitStatus(repoPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("working tree is dirty. commit or stash before running")
	}
	return nil
}

func ensureHasChanges(repoPath string) error {
	status, err := gitStatus(repoPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) == "" {
		return fmt.Errorf("no changes to commit after LLM run")
	}
	return nil
}

func invokeLLM(llm string, instructionsPath string) error {
	var cmd []string
	switch llm {
	case "codex":
		cmd = []string{"codex", "-p", instructionsPath, "--autopilot"}
	case "claude":
		cmd = []string{"claude", "-p", instructionsPath}
	case "copilot":
		cmd = []string{"copilot", "-p", instructionsPath}
	default:
		return fmt.Errorf("unsupported llm: %s", llm)
	}
	return runCmd("", cmd)
}

func gitCommitAndPush(repoPath string, message string) error {
	if err := runCmd(repoPath, []string{"git", "add", "-A"}); err != nil {
		return err
	}
	if err := ensureHasChanges(repoPath); err != nil {
		return err
	}
	if err := runCmd(repoPath, []string{"git", "commit", "-m", message}); err != nil {
		return err
	}
	if err := runCmd(repoPath, []string{"git", "push"}); err != nil {
		return err
	}
	return nil
}

func Run(cfg RunConfig) error {
	instructionsPath := filepath.Join(cfg.RepoPath, "instructions.md")

	if err := ensureClean(cfg.RepoPath); err != nil {
		return err
	}

	client := ado.NewClient(cfg.ADOOrg, cfg.ADOProject, cfg.ADOBaseURL, cfg.PAT)

	buildID := cfg.StartBuildID
	effectiveBuildDef := cfg.BuildDef

	if buildID == nil {
		baseInstructions := payload.BuildInitialInstructions(cfg.Issue, cfg.BuildDef, cfg.RepoPath)
		if err := payload.WriteInstructions(instructionsPath, baseInstructions); err != nil {
			return err
		}
		if err := invokeLLM(cfg.LLM, instructionsPath); err != nil {
			return err
		}
		if err := gitCommitAndPush(cfg.RepoPath, "Automated agent update"); err != nil {
			return err
		}
		if effectiveBuildDef == "" {
			return fmt.Errorf("missing build definition id; cannot queue a new build")
		}
		newID, err := ado.QueueBuild(client, effectiveBuildDef, "")
		if err != nil {
			return err
		}
		buildID = &newID
		time.Sleep(time.Duration(cfg.SleepSeconds) * time.Second)
	} else if effectiveBuildDef == "" {
		defID, err := ado.GetBuildDefinitionID(client, *buildID)
		if err != nil {
			return err
		}
		effectiveBuildDef = defID
	}

	for {
		status, err := ado.GetBuildStatus(client, *buildID)
		if err != nil {
			return err
		}
		if status == "completed" {
			result, err := ado.GetBuildResult(client, *buildID)
			if err != nil {
				return err
			}
			if result == "succeeded" {
				return nil
			}

			failurePayload, err := payload.BuildFailureInstructions(
				cfg.Issue,
				effectiveBuildDef,
				cfg.RepoPath,
				*buildID,
				client,
				cfg.LogMaxBytes,
			)
			if err != nil {
				return err
			}
			if err := payload.WriteInstructions(instructionsPath, failurePayload); err != nil {
				return err
			}
			if err := invokeLLM(cfg.LLM, instructionsPath); err != nil {
				return err
			}
			if err := gitCommitAndPush(cfg.RepoPath, "Automated fix for build failure"); err != nil {
				return err
			}
			if effectiveBuildDef == "" {
				return fmt.Errorf("missing build definition id; cannot queue a new build")
			}
			newID, err := ado.QueueBuild(client, effectiveBuildDef, "")
			if err != nil {
				return err
			}
			buildID = &newID
			time.Sleep(time.Duration(cfg.SleepSeconds) * time.Second)
			continue
		}
		time.Sleep(time.Duration(cfg.SleepSeconds) * time.Second)
	}
}
