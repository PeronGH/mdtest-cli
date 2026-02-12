package run

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/PeronGH/mdtest/internal/agent"
	"github.com/PeronGH/mdtest/internal/logs"
)

func TestRunReturnsSetupErrorWhenNoTests(t *testing.T) {
	root := t.TempDir()
	deps := Dependencies{
		DiscoverTests: func(string) ([]string, error) { return nil, nil },
		NextLogPath:   logs.NextLogPath,
		ParseStatus:   logs.ParseStatus,
		BuildPrompt:   func(string, string) string { return "" },
		MkdirAll:      os.MkdirAll,
		Now:           time.Now,
		Exec: func(context.Context, ExecRequest) (ExecResult, error) {
			return ExecResult{}, nil
		},
		Out: io.Discard,
	}

	_, err := Run(context.Background(), Config{Root: root, Agent: agent.ClaudeAgent}, deps)
	if err == nil {
		t.Fatal("Run returned nil error, want setup error")
	}
	var setupErr *SetupError
	if !errors.As(err, &setupErr) {
		t.Fatalf("Run error = %T, want *SetupError", err)
	}
}

func TestRunSortsTestsAndContinuesOnParseErrors(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "a.test.md"), "")
	mustWriteFile(t, filepath.Join(root, "b.test.md"), "")

	var execOrder []string
	var prompts []string
	deps := Dependencies{
		DiscoverTests: func(string) ([]string, error) {
			return []string{"b.test.md", "a.test.md"}, nil
		},
		NextLogPath: func(testAbs string, _ time.Time) (string, string, error) {
			base := strings.TrimSuffix(filepath.Base(testAbs), ".test.md")
			logDir := filepath.Join(filepath.Dir(testAbs), base+".logs")
			logAbs := filepath.Join(logDir, base+".log.md")
			return logDir, logAbs, nil
		},
		ParseStatus: func(logAbs string) (logs.Status, error) {
			if strings.HasSuffix(logAbs, "a.log.md") {
				return "", errors.New("bad log")
			}
			return logs.StatusPass, nil
		},
		BuildPrompt: func(testAbs string, logAbs string) string {
			prompt := testAbs + " -> " + logAbs
			prompts = append(prompts, prompt)
			return prompt
		},
		MkdirAll: os.MkdirAll,
		Now: func() time.Time {
			return time.Date(2026, time.February, 11, 10, 0, 0, 0, time.UTC)
		},
		Exec: func(_ context.Context, req ExecRequest) (ExecResult, error) {
			execOrder = append(execOrder, req.Argv[0])
			return ExecResult{ExitCode: 17}, nil
		},
		Out: io.Discard,
	}

	result, err := Run(context.Background(), Config{Root: root, Agent: agent.CodexAgent}, deps)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Total != 2 || result.Passed != 1 || result.Failed != 1 {
		t.Fatalf("SuiteResult = %#v, want total=2 passed=1 failed=1", result)
	}

	wantPrompts := []string{
		filepath.Join(root, "a.test.md") + " -> " + filepath.Join(root, "a.logs", "a.log.md"),
		filepath.Join(root, "b.test.md") + " -> " + filepath.Join(root, "b.logs", "b.log.md"),
	}
	if !reflect.DeepEqual(prompts, wantPrompts) {
		t.Fatalf("prompts = %#v, want %#v", prompts, wantPrompts)
	}

	wantExec := []string{"codex", "codex"}
	if !reflect.DeepEqual(execOrder, wantExec) {
		t.Fatalf("execOrder = %#v, want %#v", execOrder, wantExec)
	}
}

func TestRunReturnsSetupErrorWhenExecFails(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "only.test.md"), "")

	deps := Dependencies{
		DiscoverTests: func(string) ([]string, error) {
			return []string{"only.test.md"}, nil
		},
		NextLogPath: func(testAbs string, _ time.Time) (string, string, error) {
			logDir := filepath.Join(filepath.Dir(testAbs), "only.logs")
			return logDir, filepath.Join(logDir, "only.log.md"), nil
		},
		ParseStatus: func(string) (logs.Status, error) { return logs.StatusPass, nil },
		BuildPrompt: func(string, string) string { return "prompt" },
		MkdirAll:    os.MkdirAll,
		Now:         time.Now,
		Exec: func(context.Context, ExecRequest) (ExecResult, error) {
			return ExecResult{}, errors.New("spawn failure")
		},
		Out: io.Discard,
	}

	_, err := Run(context.Background(), Config{Root: root, Agent: agent.ClaudeAgent}, deps)
	if err == nil {
		t.Fatal("Run returned nil error, want setup error")
	}
	var setupErr *SetupError
	if !errors.As(err, &setupErr) {
		t.Fatalf("Run error = %T, want *SetupError", err)
	}
}

func TestRunUsesExplicitFilesAndSkipsDiscovery(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "b.test.md"), "")
	mustWriteFile(t, filepath.Join(root, "a.test.md"), "")

	discoverCalled := false
	var seenArgs [][]string
	deps := Dependencies{
		DiscoverTests: func(string) ([]string, error) {
			discoverCalled = true
			return nil, errors.New("discovery should not be called")
		},
		NextLogPath: func(testAbs string, _ time.Time) (string, string, error) {
			base := strings.TrimSuffix(filepath.Base(testAbs), ".test.md")
			logDir := filepath.Join(filepath.Dir(testAbs), base+".logs")
			logAbs := filepath.Join(logDir, base+".log.md")
			return logDir, logAbs, nil
		},
		ParseStatus: func(string) (logs.Status, error) {
			return logs.StatusPass, nil
		},
		BuildPrompt: func(testAbs string, logAbs string) string {
			return testAbs + " => " + logAbs
		},
		MkdirAll: os.MkdirAll,
		Now: func() time.Time {
			return time.Date(2026, time.February, 12, 10, 0, 0, 0, time.UTC)
		},
		Exec: func(_ context.Context, req ExecRequest) (ExecResult, error) {
			seenArgs = append(seenArgs, append([]string(nil), req.Argv...))
			return ExecResult{ExitCode: 0}, nil
		},
		Out: io.Discard,
	}

	result, err := Run(context.Background(), Config{
		Root:  root,
		Files: []string{"b.test.md", "a.test.md"},
		Agent: agent.CodexAgent,
	}, deps)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if discoverCalled {
		t.Fatal("DiscoverTests was called with explicit files, want it skipped")
	}
	if result.Total != 2 || result.Passed != 2 || result.Failed != 0 {
		t.Fatalf("SuiteResult = %#v, want total=2 passed=2 failed=0", result)
	}

	if len(result.Results) != 2 {
		t.Fatalf("result.Results length = %d, want 2", len(result.Results))
	}
	if result.Results[0].TestRel != "a.test.md" || result.Results[1].TestRel != "b.test.md" {
		t.Fatalf(
			"result order = [%q, %q], want lexical [a.test.md, b.test.md]",
			result.Results[0].TestRel,
			result.Results[1].TestRel,
		)
	}

	if len(seenArgs) != 2 {
		t.Fatalf("exec invocation count = %d, want 2", len(seenArgs))
	}
	if seenArgs[0][0] != "codex" || seenArgs[0][1] != "exec" {
		t.Fatalf("first argv = %#v, want codex exec ...", seenArgs[0])
	}
}
