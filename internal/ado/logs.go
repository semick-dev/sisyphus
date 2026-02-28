package ado

import (
	"fmt"
	"unicode/utf8"
)

const DefaultLogsAPIVersion = "7.1-preview.2"

func ListLogs(client *Client, buildID int, apiVersion string) ([]map[string]any, error) {
	if apiVersion == "" {
		apiVersion = DefaultLogsAPIVersion
	}
	var data map[string]any
	err := client.RequestJSON(
		"GET",
		fmt.Sprintf("/_apis/build/builds/%d/logs", buildID),
		map[string]string{"api-version": apiVersion},
		nil,
		&data,
	)
	if err != nil {
		return nil, err
	}

	items, ok := data["value"].([]any)
	if !ok {
		return []map[string]any{}, nil
	}

	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func GetLog(client *Client, buildID int, logID int, apiVersion string) (string, error) {
	if apiVersion == "" {
		apiVersion = DefaultLogsAPIVersion
	}
	return client.RequestText(
		"GET",
		fmt.Sprintf("/_apis/build/builds/%d/logs/%d", buildID, logID),
		map[string]string{"api-version": apiVersion},
	)
}

func Truncate(text string, maxBytes int) string {
	encoded := []byte(text)
	if len(encoded) <= maxBytes {
		return text
	}
	cut := encoded[:maxBytes]
	for len(cut) > 0 && !utf8.Valid(cut) {
		cut = cut[:len(cut)-1]
	}
	return string(cut)
}

func FetchFailureExcerpt(client *Client, buildID int, maxBytes int) (string, error) {
	logs, err := ListLogs(client, buildID, "")
	if err != nil {
		return "", err
	}
	if len(logs) == 0 {
		return "<no logs available>", nil
	}

	lastID, err := anyToInt(logs[len(logs)-1]["id"])
	if err != nil {
		return "", err
	}
	content, err := GetLog(client, buildID, lastID, "")
	if err != nil {
		return "", err
	}
	return Truncate(content, maxBytes), nil
}
