package push

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sisyphus/internal/ado"
	"sisyphus/internal/payload"
)

type RunConfig struct {
	BuildDef      string
	BuildURL      string
	BuildYAMLPath string
	StartBuildID  *int
	InitialPrompt string
	RepoPath      string
	CLI           string
	Branch        string
	SleepSeconds  int
	LogMaxBytes   int
	ADOOrg        string
	ADOProject    string
	ADOBaseURL    string
	PAT           string
}

type NotImplementedError struct {
	Feature string
}

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("NotImplementedError: %s", e.Feature)
}

type completedBuild struct {
	BuildID       int
	Status        string
	Result        string
	DefinitionID  string
	FailureDetail string
}

var liveConsole *consoleUI

func logStep(format string, args ...any) {
	line := fmt.Sprintf("[sisyphus] "+format, args...)
	if liveConsole != nil {
		liveConsole.Log(line)
		return
	}
	fmt.Println(line)
}

func runCmd(cwd string, cmd []string) error {
	_, _, err := runCmdCapture(cwd, cmd)
	return err
}

func runCmdCapture(cwd string, cmd []string) (string, string, error) {
	if len(cmd) == 0 {
		return "", "", fmt.Errorf("empty command")
	}
	logStep("running command: %s", summarizeCommand(cmd))
	command := exec.Command(cmd[0], cmd[1:]...)
	if cwd != "" {
		command.Dir = cwd
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("command failed: %s\nstdout:\n%s\nstderr:\n%s", strings.Join(cmd, " "), stdout.String(), stderr.String())
	}
	return stdout.String(), stderr.String(), nil
}

func summarizeCommand(cmd []string) string {
	if len(cmd) >= 4 && cmd[0] == "codex" && cmd[1] == "exec" && cmd[2] == "--full-auto" {
		return fmt.Sprintf("codex exec --full-auto <prompt:%d bytes>", len([]byte(cmd[3])))
	}
	return strings.Join(cmd, " ")
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

func repoHasChanges(repoPath string) (bool, error) {
	status, err := gitStatus(repoPath)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(status) != "", nil
}

func invokeCLI(cli string, repoPath string, prompt string) error {
	var cmd []string
	switch cli {
	case "codex":
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			return fmt.Errorf("instructions are empty")
		}
		logStep("invoking codex with prompt (%d bytes)", len([]byte(prompt)))
		if liveConsole != nil {
			liveConsole.SetContext("Codex input tail (last 40 lines):", prompt)
			liveConsole.StartCodexAnimation()
			defer liveConsole.StopCodexAnimation()
		}
		cmd = []string{"codex", "exec", "--full-auto", prompt}
	default:
		return NotImplementedError{Feature: fmt.Sprintf("cli executor %q", cli)}
	}
	stdout, stderr, err := runCmdCapture(repoPath, cmd)
	if liveConsole != nil {
		response := strings.TrimSpace(stdout)
		if strings.TrimSpace(stderr) != "" {
			if response != "" {
				response += "\n"
			}
			response += strings.TrimSpace(stderr)
		}
		if response == "" {
			response = "<no codex output captured>"
		}
		liveConsole.SetContext("Codex response tail (last 40 lines):", response)
	}
	return err
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

func commitAndPushIfChanges(repoPath string, message string) (bool, error) {
	hasChanges, err := repoHasChanges(repoPath)
	if err != nil {
		return false, err
	}
	if !hasChanges {
		return false, nil
	}
	if err := gitCommitAndPush(repoPath, message); err != nil {
		return false, err
	}
	return true, nil
}

func waitOnBuild(client *ado.Client, buildID int, sleepSeconds int) (completedBuild, error) {
	logStep("waiting for build %d to complete (poll every %ds)", buildID, sleepSeconds)
	for {
		build, err := ado.GetBuild(client, buildID, "")
		if err != nil {
			return completedBuild{}, err
		}

		status, _ := build["status"].(string)
		result, _ := build["result"].(string)
		definitionID := ado.ExtractBuildDefinitionID(build)
		failureDetail := extractSubmissionFailure(build)
		if status == "completed" {
			if result == "" {
				result = "unknown"
			}
			logStep("build %d completed with result=%s", buildID, result)
			return completedBuild{
				BuildID:       buildID,
				Status:        status,
				Result:        result,
				DefinitionID:  definitionID,
				FailureDetail: failureDetail,
			}, nil
		}

		if status == "" {
			status = "unknown"
		}
		logStep("build %d status=%s result=%s", buildID, status, result)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func extractSubmissionFailure(build map[string]any) string {
	items, ok := build["validationResults"].([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	messages := make([]string, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if msg, ok := entry["message"].(string); ok && strings.TrimSpace(msg) != "" {
			messages = append(messages, msg)
		}
	}
	return strings.Join(messages, "\n")
}

func Run(cfg RunConfig) error {
	liveConsole = newConsoleUI(cfg.RepoPath, cfg.BuildURL)
	defer func() {
		if liveConsole != nil {
			liveConsole.Close()
			liveConsole = nil
		}
	}()

	logStep("starting run (cli=%s, repo=%s)", cfg.CLI, cfg.RepoPath)
	if err := ensureClean(cfg.RepoPath); err != nil {
		return err
	}
	logStep("working tree is clean")
	if strings.TrimSpace(cfg.Branch) == "" {
		return fmt.Errorf("branch is required")
	}

	client := ado.NewClient(cfg.ADOOrg, cfg.ADOProject, cfg.ADOBaseURL, cfg.PAT)
	buildID := cfg.StartBuildID
	effectiveBuildDef := cfg.BuildDef
	effectiveBuildYAMLPath := cfg.BuildYAMLPath
	treatCurrentBuildAsFailure := buildID != nil
	iteration := 0

	if buildID == nil && strings.TrimSpace(cfg.InitialPrompt) != "" {
		logStep("running initial prompt")
		if err := invokeCLI(cfg.CLI, cfg.RepoPath, cfg.InitialPrompt); err != nil {
			return err
		}
		hasChanges, err := repoHasChanges(cfg.RepoPath)
		if err != nil {
			return err
		}
		if hasChanges {
			logStep("initial prompt produced changes; committing and pushing")
			if err := gitCommitAndPush(cfg.RepoPath, "sisyphus initial prompt"); err != nil {
				return err
			}
		} else {
			logStep("initial prompt produced no changes")
		}
	}

	for {
		var failureDetail string
		pollSeconds := cfg.SleepSeconds
		if buildID == nil {
			if effectiveBuildDef == "" {
				return fmt.Errorf("missing build definition id; cannot queue a new build")
			}
			logStep("queueing build for definition %s on branch %s", effectiveBuildDef, cfg.Branch)
			newID, err := ado.QueueBuild(client, effectiveBuildDef, cfg.Branch, "")
			if err != nil {
				failureDetail = fmt.Sprintf("Build submission failed while queueing definition %s:\n%s", effectiveBuildDef, err.Error())
				logStep("queueing failed; will run remediation prompt with submission error details")
			} else {
				buildID = &newID
				logStep("queued build id=%d", *buildID)
				pollSeconds = 10
			}
		}

		var failingBuildID *int
		if failureDetail == "" {
			result, err := waitOnBuild(client, *buildID, pollSeconds)
			if err != nil {
				return err
			}
			if result.Result == "succeeded" && !treatCurrentBuildAsFailure {
				if liveConsole != nil {
					liveConsole.ShowSuccess(result.BuildID)
				}
				return nil
			}
			if result.DefinitionID != "" {
				effectiveBuildDef = result.DefinitionID
			}
			failingBuildID = buildID
			failureDetail = result.FailureDetail
			treatCurrentBuildAsFailure = false
		}

		failurePayload, err := payload.BuildFailureInstructions(
			effectiveBuildYAMLPath,
			cfg.RepoPath,
			failingBuildID,
			client,
			cfg.LogMaxBytes,
			failureDetail,
		)
		if err != nil {
			return err
		}
		logStep("built remediation prompt (%d bytes)", len([]byte(failurePayload)))
		if err := invokeCLI(cfg.CLI, cfg.RepoPath, failurePayload); err != nil {
			return err
		}
		logStep("codex completed; checking for changes to commit")
		committed, err := commitAndPushIfChanges(cfg.RepoPath, fmt.Sprintf("sisyphus iteration %d", iteration))
		if err != nil {
			return err
		}
		if committed {
			logStep("pushed commit for iteration %d", iteration)
		} else {
			logStep("no changes produced this iteration; continuing to next build")
		}
		logStep("iteration %d complete", iteration)
		iteration++

		if effectiveBuildDef == "" {
			return fmt.Errorf("missing build definition id; cannot queue a new build")
		}
		buildID = nil
		logStep("sleeping %ds before next queue attempt", cfg.SleepSeconds)
		time.Sleep(time.Duration(cfg.SleepSeconds) * time.Second)
	}
}
