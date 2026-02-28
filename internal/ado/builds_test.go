package ado

import "testing"

func TestParseBuildURLVisualStudio(t *testing.T) {
	url := "https://sbeddall.visualstudio.com/Investigations/_build?definitionId=1"
	info, err := ParseBuildURL(url)
	if err != nil {
		t.Fatalf("ParseBuildURL() error = %v", err)
	}
	if info.Org != "sbeddall" {
		t.Fatalf("org = %q, want sbeddall", info.Org)
	}
	if info.Project != "Investigations" {
		t.Fatalf("project = %q, want Investigations", info.Project)
	}
	if info.BuildDef != "1" {
		t.Fatalf("buildDef = %q, want 1", info.BuildDef)
	}
	if info.BaseURL != "https://dev.azure.com" {
		t.Fatalf("baseURL = %q, want https://dev.azure.com", info.BaseURL)
	}
	if info.BuildID != "" {
		t.Fatalf("buildID = %q, want empty", info.BuildID)
	}
}

func TestParseBuildURLDevAzure(t *testing.T) {
	url := "https://dev.azure.com/myorg/myproject/_build?definitionId=42"
	info, err := ParseBuildURL(url)
	if err != nil {
		t.Fatalf("ParseBuildURL() error = %v", err)
	}
	if info.Org != "myorg" {
		t.Fatalf("org = %q, want myorg", info.Org)
	}
	if info.Project != "myproject" {
		t.Fatalf("project = %q, want myproject", info.Project)
	}
	if info.BuildDef != "42" {
		t.Fatalf("buildDef = %q, want 42", info.BuildDef)
	}
	if info.BaseURL != "https://dev.azure.com" {
		t.Fatalf("baseURL = %q, want https://dev.azure.com", info.BaseURL)
	}
	if info.BuildID != "" {
		t.Fatalf("buildID = %q, want empty", info.BuildID)
	}
}

func TestParseBuildURLResultsBuildID(t *testing.T) {
	url := "https://sbeddall.visualstudio.com/Investigations/_build/results?buildId=447&view=results"
	info, err := ParseBuildURL(url)
	if err != nil {
		t.Fatalf("ParseBuildURL() error = %v", err)
	}
	if info.Org != "sbeddall" {
		t.Fatalf("org = %q, want sbeddall", info.Org)
	}
	if info.Project != "Investigations" {
		t.Fatalf("project = %q, want Investigations", info.Project)
	}
	if info.BuildDef != "" {
		t.Fatalf("buildDef = %q, want empty", info.BuildDef)
	}
	if info.BaseURL != "https://dev.azure.com" {
		t.Fatalf("baseURL = %q, want https://dev.azure.com", info.BaseURL)
	}
	if info.BuildID != "447" {
		t.Fatalf("buildID = %q, want 447", info.BuildID)
	}
}
