package payload

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"sisyphus/internal/ado"
)

func TestBuildFailureInstructionsIncludesLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/logs/") {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("boom"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":7}]}`))
	}))
	defer ts.Close()

	client := ado.NewClient("myorg", "myproject", ts.URL, "token")

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
