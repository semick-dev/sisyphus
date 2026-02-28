package ado

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const DefaultBuildsAPIVersion = "7.1-preview.7"

type BuildURLInfo struct {
	Org      string
	Project  string
	BuildDef string
	BaseURL  string
	BuildID  string
}

type BuildDefinitionMetadata struct {
	ID       string
	YAMLPath string
}

func ParseBuildURL(buildURL string) (BuildURLInfo, error) {
	parsed, err := url.Parse(buildURL)
	if err != nil {
		return BuildURLInfo{}, fmt.Errorf("parse build url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return BuildURLInfo{}, fmt.Errorf("build URL must be a full URL with scheme and host")
	}

	host := parsed.Host
	var org string
	var project string
	var baseURL string

	pathParts := strings.FieldsFunc(parsed.Path, func(r rune) bool { return r == '/' })

	if strings.HasSuffix(host, ".visualstudio.com") {
		org = strings.Split(host, ".")[0]
		baseURL = "https://dev.azure.com"
		if len(pathParts) < 2 {
			return BuildURLInfo{}, fmt.Errorf("build URL path must include project name")
		}
		project = pathParts[0]
	} else {
		baseURL = parsed.Scheme + "://" + host
		if len(pathParts) > 0 {
			org = pathParts[0]
		}
		if len(pathParts) < 3 {
			return BuildURLInfo{}, fmt.Errorf("build URL path must include org and project")
		}
		project = pathParts[1]
	}

	query := parsed.Query()
	buildDef := query.Get("definitionId")
	buildID := query.Get("buildId")
	if buildDef == "" && buildID == "" {
		return BuildURLInfo{}, fmt.Errorf("build URL must include definitionId or buildId query param")
	}
	if org == "" || project == "" {
		return BuildURLInfo{}, fmt.Errorf("could not parse org/project from build URL")
	}

	return BuildURLInfo{
		Org:      org,
		Project:  project,
		BuildDef: buildDef,
		BaseURL:  baseURL,
		BuildID:  buildID,
	}, nil
}

func QueueBuild(client *Client, definition string, apiVersion string) (int, error) {
	if apiVersion == "" {
		apiVersion = DefaultBuildsAPIVersion
	}
	var data map[string]any
	err := client.RequestJSON(
		"POST",
		"/_apis/build/builds",
		map[string]string{"api-version": apiVersion},
		map[string]any{"definition": map[string]string{"id": definition}},
		&data,
	)
	if err != nil {
		return 0, err
	}
	id, err := anyToInt(data["id"])
	if err != nil {
		return 0, fmt.Errorf("queue build missing id: %w", err)
	}
	return id, nil
}

func GetBuild(client *Client, buildID int, apiVersion string) (map[string]any, error) {
	if apiVersion == "" {
		apiVersion = DefaultBuildsAPIVersion
	}
	var data map[string]any
	err := client.RequestJSON(
		"GET",
		fmt.Sprintf("/_apis/build/builds/%d", buildID),
		map[string]string{"api-version": apiVersion},
		nil,
		&data,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetBuildStatus(client *Client, buildID int) (string, error) {
	data, err := GetBuild(client, buildID, "")
	if err != nil {
		return "", err
	}
	if status, ok := data["status"].(string); ok {
		return status, nil
	}
	return "unknown", nil
}

func GetBuildResult(client *Client, buildID int) (string, error) {
	data, err := GetBuild(client, buildID, "")
	if err != nil {
		return "", err
	}
	if result, ok := data["result"].(string); ok {
		return result, nil
	}
	return "unknown", nil
}

func ExtractBuildDefinitionID(build map[string]any) string {
	definition, ok := build["definition"].(map[string]any)
	if !ok {
		return ""
	}
	switch raw := definition["id"].(type) {
	case string:
		return raw
	case float64:
		return strconv.Itoa(int(raw))
	case int:
		return strconv.Itoa(raw)
	default:
		return ""
	}
}

func GetBuildDefinitionID(client *Client, buildID int) (string, error) {
	data, err := GetBuild(client, buildID, "")
	if err != nil {
		return "", err
	}
	defID := ExtractBuildDefinitionID(data)
	if defID == "" {
		return "", fmt.Errorf("build %d does not include a definition id", buildID)
	}
	return defID, nil
}

func GetBuildDefinitionMetadata(client *Client, definitionID string, apiVersion string) (BuildDefinitionMetadata, error) {
	if definitionID == "" {
		return BuildDefinitionMetadata{}, fmt.Errorf("definition id is required")
	}
	if apiVersion == "" {
		apiVersion = DefaultBuildsAPIVersion
	}

	var data map[string]any
	err := client.RequestJSON(
		"GET",
		fmt.Sprintf("/_apis/build/definitions/%s", definitionID),
		map[string]string{"api-version": apiVersion},
		nil,
		&data,
	)
	if err != nil {
		return BuildDefinitionMetadata{}, err
	}

	id := definitionID
	if rawID, ok := data["id"]; ok {
		if parsedID, err := anyToInt(rawID); err == nil {
			id = strconv.Itoa(parsedID)
		}
	}

	yamlPath := ""
	if process, ok := data["process"].(map[string]any); ok {
		if v, ok := process["yamlFilename"].(string); ok {
			yamlPath = v
		}
	}
	if yamlPath == "" {
		if v, ok := data["yamlFilename"].(string); ok {
			yamlPath = v
		}
	}

	return BuildDefinitionMetadata{
		ID:       id,
		YAMLPath: yamlPath,
	}, nil
}

func anyToInt(v any) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case string:
		return strconv.Atoi(n)
	default:
		return 0, fmt.Errorf("unexpected numeric type %T", v)
	}
}
