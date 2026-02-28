package ado

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTruncateRespectsMaxBytes(t *testing.T) {
	text := "aaaaaaaaaa"
	truncated := Truncate(text, 5)
	if len([]byte(truncated)) > 5 {
		t.Fatalf("len(truncated) = %d, want <= 5", len([]byte(truncated)))
	}
}

func TestFetchFailureExcerptAggregatesAllLogs(t *testing.T) {
	client := NewClient("org", "proj", "https://example.invalid", "token")
	client.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body := `{"value":[{"id":1},{"id":2}]}`
			contentType := "application/json"
			if strings.HasSuffix(r.URL.Path, "/logs/1") {
				body = "first-log"
				contentType = "text/plain"
			}
			if strings.HasSuffix(r.URL.Path, "/logs/2") {
				body = "second-log"
				contentType = "text/plain"
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{contentType}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    r,
			}, nil
		}),
	}

	text, err := FetchFailureExcerpt(client, 42, 300000)
	if err != nil {
		t.Fatalf("FetchFailureExcerpt() error = %v", err)
	}
	if !strings.Contains(text, "first-log") || !strings.Contains(text, "second-log") {
		t.Fatalf("expected combined logs, got: %q", text)
	}
}
