package cli

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"

	"mdtest/internal/agent"
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
	root := NewRootCmd(stdout, stderr, lookPath)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())

		var exitErr *ExitError
		if ok := As(err, &exitErr); ok {
			return exitErr.Code
		}
		return ExitSetupError
	}

	return ExitOK
}

func NewRootCmd(stdout, stderr io.Writer, lookPath agent.LookPathFunc) *cobra.Command {
	root := &cobra.Command{
		Use:           "mdtest",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(newRunCmd(lookPath))
	return root
}

func newRunCmd(lookPath agent.LookPathFunc) *cobra.Command {
	agentFlag := string(agent.AutoMode)
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run markdown tests",
		RunE: func(*cobra.Command, []string) error {
			mode, err := agent.ParseMode(agentFlag)
			if err != nil {
				return &ExitError{Code: ExitSetupError, Err: err}
			}
			if _, err := agent.Resolve(mode, lookPath); err != nil {
				return &ExitError{Code: ExitSetupError, Err: err}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&agentFlag, "agent", string(agent.AutoMode), "Agent mode: auto, claude, or codex")
	return cmd
}

func As(err error, target interface{}) bool {
	switch t := target.(type) {
	case **ExitError:
		for err != nil {
			exitErr, ok := err.(*ExitError)
			if ok {
				*t = exitErr
				return true
			}
			unwrapper, ok := err.(interface{ Unwrap() error })
			if !ok {
				return false
			}
			err = unwrapper.Unwrap()
		}
		return false
	default:
		return false
	}
}

func DefaultLookPath(file string) (string, error) {
	return exec.LookPath(file)
}
