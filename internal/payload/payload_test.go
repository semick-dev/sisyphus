package payload

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"sisyphus/internal/ado"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestBuildFailureInstructionsIncludesLog(t *testing.T) {
	client := ado.NewClient("myorg", "myproject", "https://example.invalid", "token")
	client.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			var body string
			contentType := "application/json"
			if strings.Contains(r.URL.Path, "/logs/") {
				body = "boom"
				contentType = "text/plain"
			} else {
				body = `{"value":[{"id":7}]}`
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{contentType}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    r,
			}, nil
		}),
	}

	result, err := BuildFailureInstructions(
		"Org/repo#1",
		"99",
		filepath.Clean("."),
		123,
		client,
		10,
	)
	if err != nil {
		t.Fatalf("BuildFailureInstructions() error = %v", err)
	}
	if !strings.Contains(result, "boom") {
		t.Fatalf("result does not contain log excerpt: %q", result)
	}
	if !strings.Contains(result, "123") {
		t.Fatalf("result does not contain build id: %q", result)
	}
}
