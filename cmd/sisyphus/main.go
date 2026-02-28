package main

import (
	"bufio"
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
	Build        string
	PAT          string
	CLI          string
	Branch       string
	SleepSeconds int
	LogMaxBytes  int
}

func parseArgs() (config, error) {
	var cfg config
	fs := flag.NewFlagSet("sisyphus", flag.ContinueOnError)
	var stderr bytes.Buffer
	fs.SetOutput(&stderr)

	fs.StringVar(&cfg.Build, "build", "", "ADO build definition URL or build results URL.")
	fs.StringVar(&cfg.PAT, "pat", "", "ADO PAT token. Optionally sourced from ADO_PAT environment variable. Requires Azure DevOps Build scope with Read and Execute permissions.")
	fs.StringVar(&cfg.CLI, "cli", "codex", "CLI executor to invoke for autopilot.")
	fs.IntVar(&cfg.SleepSeconds, "sleep-seconds", 30, "Seconds to sleep between loop iterations and for polling a provided starting buildId. Newly queued builds are polled every 10 seconds.")
	fs.IntVar(&cfg.LogMaxBytes, "log-max-bytes", 300000, "Max bytes of log content to attach to instructions.")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cfg, fmt.Errorf("%w\n%s", err, stderr.String())
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
	if strings.TrimSpace(cfg.CLI) == "" {
		return cfg, fmt.Errorf("--cli must not be empty")
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

func promptForInitialPrompt() (string, error) {
	fmt.Fprintln(os.Stdout, "Do you have an initial prompt you'd like to start with?")
	fmt.Fprintln(os.Stdout, "Enter prompt text, then submit an empty line to continue:")

	scanner := bufio.NewScanner(os.Stdin)
	lines := make([]string, 0, 8)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
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
	cfg.Branch = branch

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
	var initialPrompt string
	if startBuildID == nil {
		initialPrompt, err = promptForInitialPrompt()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	client := ado.NewClient(info.Org, info.Project, info.BaseURL, cfg.PAT)
	effectiveBuildDef := info.BuildDef
	if effectiveBuildDef == "" && startBuildID != nil {
		defID, err := ado.GetBuildDefinitionID(client, *startBuildID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		effectiveBuildDef = defID
	}

	var buildYAMLPath string
	if effectiveBuildDef != "" {
		meta, err := ado.GetBuildDefinitionMetadata(client, effectiveBuildDef, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		effectiveBuildDef = meta.ID
		buildYAMLPath = meta.YAMLPath
	}

	err = push.Run(push.RunConfig{
		BuildDef:      effectiveBuildDef,
		BuildURL:      cfg.Build,
		BuildYAMLPath: buildYAMLPath,
		StartBuildID:  startBuildID,
		InitialPrompt: initialPrompt,
		RepoPath:      repoPath,
		CLI:           cfg.CLI,
		Branch:        cfg.Branch,
		SleepSeconds:  cfg.SleepSeconds,
		LogMaxBytes:   cfg.LogMaxBytes,
		ADOOrg:        info.Org,
		ADOProject:    info.Project,
		ADOBaseURL:    info.BaseURL,
		PAT:           cfg.PAT,
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
