package cli

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestExecuteRejectsInvalidAgentFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run", "--agent", "bogus"}, &stdout, &stderr, func(string) (string, error) {
		return "", exec.ErrNotFound
	})

	if code != 2 {
		t.Fatalf("Execute exit code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want validation error output")
	}
}

func TestExecuteRunUsesDefaultAutoAndPassesWhenClaudeExists(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run"}, &stdout, &stderr, func(file string) (string, error) {
		if file == "claude" {
			return "/usr/bin/claude", nil
		}
		return "", exec.ErrNotFound
	})

	if code != 0 {
		t.Fatalf("Execute exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
}

func TestExecuteRunFailsSetupWhenRequestedAgentMissing(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run", "--agent", "codex"}, &stdout, &stderr, func(string) (string, error) {
		return "", exec.ErrNotFound
	})

	if code != 2 {
		t.Fatalf("Execute exit code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want setup error output")
	}
}
