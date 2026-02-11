package cli

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"testing"

	"mdtest/internal/run"
)

func TestExecuteRejectsInvalidAgentFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeWithDeps(
		[]string{"run", "--agent", "bogus"},
		&stdout,
		&stderr,
		func(string) (string, error) { return "", exec.ErrNotFound },
		func(context.Context, run.Config) (run.SuiteResult, error) { return run.SuiteResult{}, nil },
	)

	if code != 2 {
		t.Fatalf("Execute exit code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want validation error output")
	}
}

func TestExecuteRunUsesDefaultAutoAndReturnsPassCode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeWithDeps(
		[]string{"run"},
		&stdout,
		&stderr,
		func(file string) (string, error) {
			if file == "claude" {
				return "/usr/bin/claude", nil
			}
			return "", exec.ErrNotFound
		},
		func(context.Context, run.Config) (run.SuiteResult, error) {
			return run.SuiteResult{Total: 2, Passed: 2, Failed: 0}, nil
		},
	)

	if code != 0 {
		t.Fatalf("Execute exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
}

func TestExecuteRunFailsSetupWhenRequestedAgentMissing(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeWithDeps(
		[]string{"run", "--agent", "codex"},
		&stdout,
		&stderr,
		func(string) (string, error) { return "", exec.ErrNotFound },
		func(context.Context, run.Config) (run.SuiteResult, error) { return run.SuiteResult{}, nil },
	)

	if code != 2 {
		t.Fatalf("Execute exit code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want setup error output")
	}
}

func TestExecuteRunReturnsFailureCodeWhenSuiteHasFailures(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeWithDeps(
		[]string{"run"},
		&stdout,
		&stderr,
		func(file string) (string, error) {
			if file == "claude" {
				return "/usr/bin/claude", nil
			}
			return "", exec.ErrNotFound
		},
		func(context.Context, run.Config) (run.SuiteResult, error) {
			return run.SuiteResult{Total: 3, Passed: 2, Failed: 1}, nil
		},
	)

	if code != 1 {
		t.Fatalf("Execute exit code = %d, want 1", code)
	}
}

func TestExecuteRunReturnsSetupCodeWhenRunnerErrors(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := executeWithDeps(
		[]string{"run"},
		&stdout,
		&stderr,
		func(file string) (string, error) {
			if file == "claude" {
				return "/usr/bin/claude", nil
			}
			return "", exec.ErrNotFound
		},
		func(context.Context, run.Config) (run.SuiteResult, error) {
			return run.SuiteResult{}, &run.SetupError{Err: errors.New("no tests")}
		},
	)

	if code != 2 {
		t.Fatalf("Execute exit code = %d, want 2", code)
	}
}
