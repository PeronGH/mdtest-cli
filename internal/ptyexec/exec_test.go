package ptyexec

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestRunRequiresArgv(t *testing.T) {
	_, err := Run(context.Background(), Request{RootAbs: t.TempDir()})
	if err == nil {
		t.Fatal("Run returned nil error, want failure")
	}
}

func TestRunReturnsChildExitCode(t *testing.T) {
	stdinR, stdinW, stdoutR, stdoutW := newPipes(t)
	_ = stdinW.Close()
	defer stdoutR.Close()

	got, err := runWithConfig(context.Background(), Request{
		RootAbs: t.TempDir(),
		Argv:    []string{"sh", "-c", "exit 7"},
	}, runtimeConfig{
		stdin:        stdinR,
		stdout:       stdoutW,
		signalSource: make(chan os.Signal),
		isTerminal:   func(int) bool { return false },
	})
	_ = stdoutW.Close()

	if err != nil {
		t.Fatalf("runWithConfig returned error: %v", err)
	}
	if got.ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", got.ExitCode)
	}
}

func TestRunForwardsTerminationSignalsToChildSession(t *testing.T) {
	stdinR, stdinW, stdoutR, stdoutW := newPipes(t)
	_ = stdinW.Close()
	defer stdoutR.Close()

	sigCh := make(chan os.Signal, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	type runOutcome struct {
		res Result
		err error
	}
	done := make(chan runOutcome, 1)
	go func() {
		res, err := runWithConfig(ctx, Request{
			RootAbs: t.TempDir(),
			Argv:    []string{"sh", "-c", "trap 'exit 0' TERM INT HUP QUIT; while :; do sleep 1; done"},
		}, runtimeConfig{
			stdin:        stdinR,
			stdout:       stdoutW,
			signalSource: sigCh,
			isTerminal:   func(int) bool { return false },
		})
		done <- runOutcome{res: res, err: err}
	}()

	time.Sleep(250 * time.Millisecond)
	sigCh <- syscall.SIGTERM

	out := <-done
	_ = stdoutW.Close()
	if out.err != nil {
		t.Fatalf("runWithConfig returned error: %v", out.err)
	}
	if out.res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", out.res.ExitCode)
	}
}

func TestRunReturnsErrorWhenTerminalRestoreFails(t *testing.T) {
	stdinR, stdinW, stdoutR, stdoutW := newPipes(t)
	_ = stdinW.Close()
	defer stdoutR.Close()

	_, err := runWithConfig(context.Background(), Request{
		RootAbs: t.TempDir(),
		Argv:    []string{"sh", "-c", "exit 0"},
	}, runtimeConfig{
		stdin:        stdinR,
		stdout:       stdoutW,
		signalSource: make(chan os.Signal),
		isTerminal:   func(int) bool { return true },
		makeRaw: func(int) (*term.State, error) {
			return &term.State{}, nil
		},
		restore: func(int, *term.State) error {
			return errors.New("restore failed")
		},
	})
	_ = stdoutW.Close()

	if err == nil {
		t.Fatal("runWithConfig returned nil error, want failure")
	}
}

func newPipes(t *testing.T) (*os.File, *os.File, *os.File, *os.File) {
	t.Helper()
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe (stdin): %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe (stdout): %v", err)
	}
	t.Cleanup(func() {
		_ = stdinR.Close()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stdoutW.Close()
	})
	return stdinR, stdinW, stdoutR, stdoutW
}
