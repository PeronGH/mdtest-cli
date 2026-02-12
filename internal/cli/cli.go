package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/PeronGH/mdtest/internal/agent"
	"github.com/PeronGH/mdtest/internal/procexec"
	"github.com/PeronGH/mdtest/internal/run"
)

const (
	ExitOK         = 0
	ExitFailed     = 1
	ExitSetupError = 2
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

func Execute(args []string, stdout, stderr io.Writer, lookPath agent.LookPathFunc) int {
	return executeWithDeps(args, stdout, stderr, lookPath, defaultRunSuite(stdout))
}

type RunSuiteFunc func(ctx context.Context, cfg run.Config) (run.SuiteResult, error)

func executeWithDeps(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	lookPath agent.LookPathFunc,
	runSuite RunSuiteFunc,
) int {
	root := NewRootCmd(stdout, stderr, lookPath, runSuite)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())

		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}
		return ExitSetupError
	}

	return ExitOK
}

func NewRootCmd(stdout, stderr io.Writer, lookPath agent.LookPathFunc, runSuite RunSuiteFunc) *cobra.Command {
	root := &cobra.Command{
		Use:           "mdtest",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(newRunCmd(lookPath, runSuite))
	return root
}

func newRunCmd(lookPath agent.LookPathFunc, runSuite RunSuiteFunc) *cobra.Command {
	agentFlag := string(agent.AutoMode)
	dirFlag := "."
	var fileFlags []string
	interactiveFlag := false
	dangerousFlag := false
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run markdown tests",
		RunE: func(*cobra.Command, []string) error {
			mode, err := agent.ParseMode(agentFlag)
			if err != nil {
				return &ExitError{Code: ExitSetupError, Err: err}
			}
			resolved, err := agent.Resolve(mode, lookPath)
			if err != nil {
				return &ExitError{Code: ExitSetupError, Err: err}
			}

			suite, err := runSuite(context.Background(), run.Config{
				Root:                       dirFlag,
				Files:                      fileFlags,
				Agent:                      resolved,
				Interactive:                interactiveFlag,
				DangerouslyAllowAllActions: dangerousFlag,
			})
			if err != nil {
				var setupErr *run.SetupError
				if errors.As(err, &setupErr) {
					return &ExitError{Code: ExitSetupError, Err: setupErr}
				}
				return &ExitError{Code: ExitSetupError, Err: err}
			}
			if suite.Failed > 0 {
				return &ExitError{Code: ExitFailed, Err: fmt.Errorf("%d test(s) failed", suite.Failed)}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentFlag, "agent", "a", string(agent.AutoMode), "Agent mode: auto, claude, or codex")
	cmd.Flags().StringVarP(&dirFlag, "dir", "d", ".", "Suite root directory")
	cmd.Flags().StringArrayVarP(&fileFlags, "file", "f", nil, "Target test file (repeatable)")
	cmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Run agent in interactive mode")
	cmd.Flags().BoolVarP(&dangerousFlag, "dangerously-allow-all-actions", "A", false, "Disable agent safety approvals/sandboxing")
	return cmd
}

func DefaultLookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func defaultRunSuite(out io.Writer) RunSuiteFunc {
	return func(ctx context.Context, cfg run.Config) (run.SuiteResult, error) {
		deps := run.DefaultDependencies(out, func(ctx context.Context, req run.ExecRequest) (run.ExecResult, error) {
			execResult, err := procexec.Run(ctx, procexec.Request{
				RootAbs:     req.RootAbs,
				Argv:        req.Argv,
				Interactive: req.Interactive,
			})
			return run.ExecResult{ExitCode: execResult.ExitCode}, err
		})
		return run.Run(ctx, cfg, deps)
	}
}
