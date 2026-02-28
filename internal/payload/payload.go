package payload

import (
	"fmt"
	"os"
	"path/filepath"

	"sisyphus/internal/ado"
)

const baseTemplate = `
Issue: %s

Instructions:
- We are attempting to resolve a failing build at
`

const failureTemplate = `
Build Failure Detected
Build ID: %d

Failed Log Excerpt (truncated):
%s
`

func BuildInitialInstructions(issue string, buildDef string, repoPath string) string {
	_ = buildDef
	_ = repoPath
	return fmt.Sprintf(baseTemplate, issue)
}

func BuildFailureInstructions(issue string, buildDef string, repoPath string, buildID int, client *ado.Client, logMaxBytes int) (string, error) {
	base := BuildInitialInstructions(issue, buildDef, repoPath)
	logExcerpt, err := ado.FetchFailureExcerpt(client, buildID, logMaxBytes)
	if err != nil {
		return "", err
	}
	return base + fmt.Sprintf(failureTemplate, buildID, logExcerpt), nil
}

func WriteInstructions(path string, content string) error {
	return os.WriteFile(filepath.Clean(path), []byte(content), 0o644)
}
