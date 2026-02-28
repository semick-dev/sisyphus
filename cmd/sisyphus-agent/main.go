package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"sisyphus/internal/ado"
	"sisyphus/internal/push"
)

type config struct {
	Issue       string
	Build       string
	PAT         string
	LLM         string
	SleepSeconds int
	LogMaxBytes int
}

func parseArgs() (config, error) {
	var cfg config
	fs := flag.NewFlagSet("sisyphus-agent", flag.ContinueOnError)
	var stderr bytes.Buffer
	fs.SetOutput(&stderr)

	fs.StringVar(&cfg.Issue, "issue", "", "Issue identifier (Org/repo#xxx) or starting prompt.")
	fs.StringVar(&cfg.Build, "build", "", "ADO build definition URL or build results URL.")
	fs.StringVar(&cfg.PAT, "pat", "", "ADO PAT token. Optionally sourced from ADO_PAT environment variable.")
	fs.StringVar(&cfg.LLM, "llm", "codex", "LLM CLI to invoke for autopilot (codex|claude|copilot).")
	fs.IntVar(&cfg.SleepSeconds, "sleep-seconds", 30, "Seconds to sleep between build status checks.")
	fs.IntVar(&cfg.LogMaxBytes, "log-max-bytes", 200000, "Max bytes of log content to attach to instructions.")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cfg, fmt.Errorf("%w\n%s", err, stderr.String())
	}
	if cfg.Issue == "" {
		return cfg, fmt.Errorf("--issue is required")
	}
	if cfg.Build == "" {
		return cfg, fmt.Errorf("--build is required")
	}
	if cfg.PAT == "" {
		cfg.PAT = os.Getenv("ADO_PAT")
	}
	if cfg.PAT == "" {
		return cfg, fmt.Errorf("an ADO PAT must be provided in --pat or ADO_PAT environment variable")
	}
	switch cfg.LLM {
	case "codex", "claude", "copilot":
	default:
		return cfg, fmt.Errorf("--llm must be one of codex, claude, copilot")
	}

	return cfg, nil
}

func currentBranch(repoPath string) (string, error) {
	command := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	command.Dir = repoPath
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("failed to resolve git branch: %s", strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func run() int {
	cfg, err := parseArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	repoPath, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
		fmt.Fprintf(os.Stderr, "Current directory is not a git repo: %s\n", repoPath)
		return 2
	}

	branch, err := currentBranch(repoPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if branch == "main" {
		fmt.Fprintln(os.Stderr, "Refusing to run on branch 'main'. Create a working branch first.")
		return 2
	}
	if branch == "HEAD" {
		fmt.Fprintln(os.Stderr, "Refusing to run on detached HEAD. Create a working branch first.")
		return 2
	}

	info, err := ado.ParseBuildURL(cfg.Build)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid --build URL: %v\n", err)
		return 2
	}

	var startBuildID *int
	if info.BuildID != "" {
		id, err := strconv.Atoi(info.BuildID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --build URL: buildId must be an integer: %v\n", err)
			return 2
		}
		startBuildID = &id
	}

	err = push.Run(push.RunConfig{
		Issue:        cfg.Issue,
		BuildDef:     info.BuildDef,
		StartBuildID: startBuildID,
		RepoPath:     repoPath,
		LLM:          cfg.LLM,
		SleepSeconds: cfg.SleepSeconds,
		LogMaxBytes:  cfg.LogMaxBytes,
		ADOOrg:       info.Org,
		ADOProject:   info.Project,
		ADOBaseURL:   info.BaseURL,
		PAT:          cfg.PAT,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
