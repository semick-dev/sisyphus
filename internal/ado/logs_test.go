package ado

import "testing"

func TestTruncateRespectsMaxBytes(t *testing.T) {
	text := "aaaaaaaaaa"
	truncated := Truncate(text, 5)
	if len([]byte(truncated)) > 5 {
		t.Fatalf("len(truncated) = %d, want <= 5", len([]byte(truncated)))
	}
}
