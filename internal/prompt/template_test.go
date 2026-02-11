package prompt

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderIncludesRequiredFacts(t *testing.T) {
	testAbs := filepath.Join("/tmp", "suite", "checkout.test.md")
	logAbs := filepath.Join("/tmp", "suite", "checkout.logs", "2026-02-11T10-00-00Z.log.md")

	got := Render(testAbs, logAbs)

	checks := []string{
		"step by step",
		testAbs,
		logAbs,
		"status: pass|fail",
	}
	for _, want := range checks {
		if !strings.Contains(strings.ToLower(got), strings.ToLower(want)) {
			t.Fatalf("Render output missing %q\nPrompt:\n%s", want, got)
		}
	}
}
