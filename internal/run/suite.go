package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/PeronGH/mdtest/internal/agent"
	"github.com/PeronGH/mdtest/internal/logs"
	"github.com/PeronGH/mdtest/internal/prompt"
)

type TestStatus string

const (
	TestPass TestStatus = "pass"
	TestFail TestStatus = "fail"
)

type TestCase struct {
	RootAbs   string
	TestAbs   string
	TestRel   string
	LogDirAbs string
	LogAbs    string
}

type TestResult struct {
	TestRel string
	LogAbs  string
	Status  TestStatus
	Reason  string
}

type SuiteResult struct {
	Total   int
	Passed  int
	Failed  int
	Results []TestResult
}

type Config struct {
	Root                       string
	Files                      []string
	Agent                      agent.Name
	Interactive                bool
	DangerouslyAllowAllActions bool
}

type ExecRequest struct {
	RootAbs string
	Argv    []string
}

type ExecResult struct {
	ExitCode int
}

type ExecFunc func(ctx context.Context, req ExecRequest) (ExecResult, error)

type Dependencies struct {
	DiscoverTests func(rootAbs string) ([]string, error)
	NextLogPath   func(testAbs string, at time.Time) (string, string, error)
	ParseStatus   func(path string) (logs.Status, error)
	BuildPrompt   func(testAbs string, logAbs string) string
	MkdirAll      func(path string, perm os.FileMode) error
	Now           func() time.Time
	Exec          ExecFunc
	Out           io.Writer
}

type SetupError struct {
	Err error
}

func (e *SetupError) Error() string {
	return e.Err.Error()
}

func (e *SetupError) Unwrap() error {
	return e.Err
}

func DefaultDependencies(out io.Writer, execFn ExecFunc) Dependencies {
	return Dependencies{
		DiscoverTests: DiscoverTests,
		NextLogPath:   logs.NextLogPath,
		ParseStatus:   logs.ParseStatus,
		BuildPrompt:   prompt.Render,
		MkdirAll:      os.MkdirAll,
		Now:           time.Now,
		Exec:          execFn,
		Out:           out,
	}
}

func Run(ctx context.Context, cfg Config, deps Dependencies) (SuiteResult, error) {
	deps = fillDefaults(deps)

	root := cfg.Root
	if root == "" {
		root = "."
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return SuiteResult{}, &SetupError{Err: fmt.Errorf("resolve root: %w", err)}
	}

	tests, err := deps.DiscoverTests(rootAbs)
	if err != nil {
		return SuiteResult{}, &SetupError{Err: fmt.Errorf("discover tests: %w", err)}
	}
	sort.Strings(tests)
	if len(tests) == 0 {
		return SuiteResult{}, &SetupError{Err: fmt.Errorf("no tests found under %s", rootAbs)}
	}

	suite := SuiteResult{
		Total:   len(tests),
		Results: make([]TestResult, 0, len(tests)),
	}
	for _, testRel := range tests {
		testAbs := filepath.Join(rootAbs, filepath.FromSlash(testRel))
		logDir, logAbs, err := deps.NextLogPath(testAbs, deps.Now().UTC())
		if err != nil {
			return SuiteResult{}, &SetupError{Err: fmt.Errorf("next log path for %s: %w", testRel, err)}
		}
		if err := deps.MkdirAll(logDir, 0o755); err != nil {
			return SuiteResult{}, &SetupError{Err: fmt.Errorf("create log dir for %s: %w", testRel, err)}
		}

		promptText := deps.BuildPrompt(testAbs, logAbs)
		argv, err := agent.CommandArgs(cfg.Agent, promptText, agent.CommandOptions{
			Interactive:                cfg.Interactive,
			DangerouslyAllowAllActions: cfg.DangerouslyAllowAllActions,
		})
		if err != nil {
			return SuiteResult{}, &SetupError{Err: fmt.Errorf("build command for %s: %w", testRel, err)}
		}

		execResult, err := deps.Exec(ctx, ExecRequest{
			RootAbs: rootAbs,
			Argv:    argv,
		})
		if err != nil {
			return SuiteResult{}, &SetupError{Err: fmt.Errorf("execute %s: %w", testRel, err)}
		}

		status, parseErr := deps.ParseStatus(logAbs)
		result := TestResult{
			TestRel: testRel,
			LogAbs:  logAbs,
		}
		if parseErr != nil {
			result.Status = TestFail
			result.Reason = fmt.Sprintf("log parse error: %v (agent exit code %d)", parseErr, execResult.ExitCode)
			suite.Failed++
		} else if status == logs.StatusPass {
			result.Status = TestPass
			suite.Passed++
		} else {
			result.Status = TestFail
			result.Reason = fmt.Sprintf("status=%s (agent exit code %d)", status, execResult.ExitCode)
			suite.Failed++
		}
		suite.Results = append(suite.Results, result)
	}

	_, _ = fmt.Fprintf(deps.Out, "Total: %d, Passed: %d, Failed: %d\n", suite.Total, suite.Passed, suite.Failed)
	return suite, nil
}

func fillDefaults(deps Dependencies) Dependencies {
	if deps.DiscoverTests == nil {
		deps.DiscoverTests = DiscoverTests
	}
	if deps.NextLogPath == nil {
		deps.NextLogPath = logs.NextLogPath
	}
	if deps.ParseStatus == nil {
		deps.ParseStatus = logs.ParseStatus
	}
	if deps.BuildPrompt == nil {
		deps.BuildPrompt = prompt.Render
	}
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.Exec == nil {
		deps.Exec = func(context.Context, ExecRequest) (ExecResult, error) {
			return ExecResult{}, fmt.Errorf("executor is not configured")
		}
	}
	if deps.Out == nil {
		deps.Out = io.Discard
	}
	return deps
}
