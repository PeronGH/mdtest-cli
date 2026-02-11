package agent

import (
	"errors"
	"os/exec"
	"testing"
)

func TestResolveAutoPrefersClaude(t *testing.T) {
	lookPath := func(file string) (string, error) {
		switch file {
		case "claude":
			return "/usr/bin/claude", nil
		case "codex":
			return "/usr/bin/codex", nil
		default:
			return "", exec.ErrNotFound
		}
	}

	got, err := Resolve(AutoMode, lookPath)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != ClaudeAgent {
		t.Fatalf("Resolve returned %q, want %q", got, ClaudeAgent)
	}
}

func TestResolveAutoFallsBackToCodex(t *testing.T) {
	lookPath := func(file string) (string, error) {
		switch file {
		case "claude":
			return "", exec.ErrNotFound
		case "codex":
			return "/usr/bin/codex", nil
		default:
			return "", exec.ErrNotFound
		}
	}

	got, err := Resolve(AutoMode, lookPath)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != CodexAgent {
		t.Fatalf("Resolve returned %q, want %q", got, CodexAgent)
	}
}

func TestResolveAutoErrorsWhenNoAgentIsAvailable(t *testing.T) {
	lookPath := func(file string) (string, error) {
		return "", exec.ErrNotFound
	}

	_, err := Resolve(AutoMode, lookPath)
	if err == nil {
		t.Fatal("Resolve returned nil error, want failure")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("Resolve error = %T, want *NotFoundError", err)
	}
}

func TestResolveExplicitModeRequiresThatBinary(t *testing.T) {
	lookPath := func(file string) (string, error) {
		if file == "claude" {
			return "", exec.ErrNotFound
		}
		return "/usr/bin/codex", nil
	}

	_, err := Resolve(ClaudeMode, lookPath)
	if err == nil {
		t.Fatal("Resolve returned nil error, want failure")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("Resolve error = %T, want *NotFoundError", err)
	}
	if notFound.Agent != ClaudeAgent {
		t.Fatalf("NotFoundError.Agent = %q, want %q", notFound.Agent, ClaudeAgent)
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Mode
		wantErr bool
	}{
		{name: "auto", raw: "auto", want: AutoMode},
		{name: "claude", raw: "claude", want: ClaudeMode},
		{name: "codex", raw: "codex", want: CodexMode},
		{name: "trim and case", raw: "  CoDeX ", want: CodexMode},
		{name: "invalid", raw: "other", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseMode returned nil error, want failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseMode = %q, want %q", got, tt.want)
			}
		})
	}
}
