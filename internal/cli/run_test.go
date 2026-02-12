package cli

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"reflect"
	"testing"

	"github.com/PeronGH/mdtest-cli/internal/agent"
	"github.com/PeronGH/mdtest-cli/internal/run"
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
	var gotCfg run.Config

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
		func(_ context.Context, cfg run.Config) (run.SuiteResult, error) {
			gotCfg = cfg
			return run.SuiteResult{Total: 2, Passed: 2, Failed: 0}, nil
		},
	)

	if code != 0 {
		t.Fatalf("Execute exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	wantCfg := run.Config{
		Root:  ".",
		Agent: agent.ClaudeAgent,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Fatalf("run config = %#v, want %#v", gotCfg, wantCfg)
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

func TestExecuteRunParsesLongFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var gotCfg run.Config

	code := executeWithDeps(
		[]string{
			"run",
			"--agent", "codex",
			"--dir", "tests/smoke",
			"--interactive",
			"--dangerously-allow-all-actions",
			"a.test.md",
			"nested/b.test.md",
		},
		&stdout,
		&stderr,
		func(file string) (string, error) {
			if file == "codex" {
				return "/usr/bin/codex", nil
			}
			return "", exec.ErrNotFound
		},
		func(_ context.Context, cfg run.Config) (run.SuiteResult, error) {
			gotCfg = cfg
			return run.SuiteResult{Total: 1, Passed: 1, Failed: 0}, nil
		},
	)

	if code != 0 {
		t.Fatalf("Execute exit code = %d, want 0; stderr=%q", code, stderr.String())
	}

	wantCfg := run.Config{
		Root:                       "tests/smoke",
		Files:                      []string{"a.test.md", "nested/b.test.md"},
		Agent:                      agent.CodexAgent,
		Interactive:                true,
		DangerouslyAllowAllActions: true,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Fatalf("run config = %#v, want %#v", gotCfg, wantCfg)
	}
}

func TestExecuteRunParsesShortFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var gotCfg run.Config

	code := executeWithDeps(
		[]string{
			"run",
			"-a", "codex",
			"-d", "tests/smoke",
			"-i",
			"-A",
			"a.test.md",
			"nested/b.test.md",
		},
		&stdout,
		&stderr,
		func(file string) (string, error) {
			if file == "codex" {
				return "/usr/bin/codex", nil
			}
			return "", exec.ErrNotFound
		},
		func(_ context.Context, cfg run.Config) (run.SuiteResult, error) {
			gotCfg = cfg
			return run.SuiteResult{Total: 1, Passed: 1, Failed: 0}, nil
		},
	)

	if code != 0 {
		t.Fatalf("Execute exit code = %d, want 0; stderr=%q", code, stderr.String())
	}

	wantCfg := run.Config{
		Root:                       "tests/smoke",
		Files:                      []string{"a.test.md", "nested/b.test.md"},
		Agent:                      agent.CodexAgent,
		Interactive:                true,
		DangerouslyAllowAllActions: true,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Fatalf("run config = %#v, want %#v", gotCfg, wantCfg)
	}
}
