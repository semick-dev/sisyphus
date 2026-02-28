package payload

import (
	"fmt"
	"strings"

	"sisyphus/internal/ado"
)

const baseTemplate = "Instructions:\n" +
	"- You are attempting to resolve a failing build located at {failing_build_url}\n\n" +
	"Here is a log exerpt from the failing build:\n\n" +
	"```\n" +
	"{logExcerpt}\n" +
	"```\n\n" +
	"- The build definition is based upon the yaml at '{build_definition_yaml_path}'.\n" +
	"- Make any local changes necessary that you feel will address the failure.\n" +
	"- If absolutely necessary, ADO_PAT is available in environment if you need to make limited requests of ADO with an identity that can interrogate the build system over REST calls. Do not abuse this." +
	"- State your conclusions, then exit this copilot prompt with 0 if you feel you succeeded, otherwise exit with 1\n"

func BuildFailureInstructions(buildYAMLPath string, repoPath string, buildID *int, client *ado.Client, logMaxBytes int, failureDetail string) (string, error) {
	_ = repoPath
	buildURL := "<queue submission failed>"
	logExcerpt := ""
	failureDetail = strings.TrimSpace(failureDetail)
	if buildID != nil {
		buildURL = fmt.Sprintf(
			"%s/%s/%s/_build/results?buildId=%d&view=results",
			strings.TrimRight(client.BaseURL, "/"),
			client.Org,
			client.Project,
			*buildID,
		)
		fetched, err := ado.FetchFailureExcerpt(client, *buildID, logMaxBytes)
		if err != nil {
			if failureDetail == "" {
				return "", err
			}
			logExcerpt = failureDetail
			failureDetail = fmt.Sprintf("Failed to fetch build logs: %v", err)
		} else {
			logExcerpt = fetched
		}
	} else {
		logExcerpt = failureDetail
		failureDetail = ""
	}
	if failureDetail != "" && logExcerpt != "" {
		logExcerpt = logExcerpt + "\n\nAdditional failure detail:\n" + failureDetail
	}
	if logExcerpt == "" {
		logExcerpt = "<no failure details available>"
	}
	yamlPath := buildYAMLPath
	if yamlPath == "" {
		yamlPath = "<unknown yaml path>"
	}

	out := strings.NewReplacer(
		"{failing_build_url}", buildURL,
		"{logExcerpt}", logExcerpt,
		"{build_definition_yaml_path}", yamlPath,
	).Replace(baseTemplate)
	return out, nil
}
