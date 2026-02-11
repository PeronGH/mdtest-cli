package ptyexec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type Request struct {
	RootAbs string
	Argv    []string
}

type Result struct {
	ExitCode int
}

func Run(ctx context.Context, req Request) (Result, error) {
	return runWithConfig(ctx, req, runtimeConfig{})
}

type runtimeConfig struct {
	stdin        *os.File
	stdout       *os.File
	ptyStart     func(cmd *exec.Cmd) (*os.File, error)
	copyStream   func(dst io.Writer, src io.Reader) (int64, error)
	inheritSize  func(pty *os.File, tty *os.File) error
	isTerminal   func(fd int) bool
	makeRaw      func(fd int) (*term.State, error)
	restore      func(fd int, oldState *term.State) error
	kill         func(pid int, sig syscall.Signal) error
	signalSource <-chan os.Signal
	notify       func(c chan<- os.Signal, sig ...os.Signal)
	stopNotify   func(c chan<- os.Signal)
}

func runWithConfig(ctx context.Context, req Request, cfg runtimeConfig) (res Result, retErr error) {
	cfg = fillDefaults(cfg)

	if len(req.Argv) == 0 {
		return Result{}, fmt.Errorf("argv is empty")
	}

	cmd := exec.CommandContext(ctx, req.Argv[0], req.Argv[1:]...)
	cmd.Dir = req.RootAbs
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	ptmx, err := cfg.ptyStart(cmd)
	if err != nil {
		return Result{}, fmt.Errorf("start pty command: %w", err)
	}
	defer func() {
		_ = ptmx.Close()
	}()

	restoreTerminal, err := prepareTerminal(cfg, cfg.stdin, ptmx)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		if restoreTerminal == nil {
			return
		}
		if err := restoreTerminal(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	stopSignals, signalErr, waitSignalLoop := startSignalForwarder(cfg, ptmx, cfg.stdin, cmd.Process.Pid)
	defer func() {
		close(stopSignals)
		waitSignalLoop()
		if retErr == nil {
			select {
			case err := <-signalErr:
				retErr = err
			default:
			}
		}
	}()

	go func() {
		_, _ = cfg.copyStream(ptmx, cfg.stdin)
	}()

	stdoutDone := make(chan struct{})
	go func() {
		_, _ = cfg.copyStream(cfg.stdout, ptmx)
		close(stdoutDone)
	}()

	waitErr := cmd.Wait()
	_ = ptmx.Close()
	<-stdoutDone

	if waitErr != nil {
		var exitErr *exec.ExitError
		if ok := asExitError(waitErr, &exitErr); ok {
			return Result{ExitCode: exitErr.ExitCode()}, nil
		}
		return Result{}, waitErr
	}
	return Result{ExitCode: 0}, nil
}

func fillDefaults(cfg runtimeConfig) runtimeConfig {
	if cfg.stdin == nil {
		cfg.stdin = os.Stdin
	}
	if cfg.stdout == nil {
		cfg.stdout = os.Stdout
	}
	if cfg.ptyStart == nil {
		cfg.ptyStart = pty.Start
	}
	if cfg.copyStream == nil {
		cfg.copyStream = io.Copy
	}
	if cfg.inheritSize == nil {
		cfg.inheritSize = pty.InheritSize
	}
	if cfg.isTerminal == nil {
		cfg.isTerminal = term.IsTerminal
	}
	if cfg.makeRaw == nil {
		cfg.makeRaw = term.MakeRaw
	}
	if cfg.restore == nil {
		cfg.restore = term.Restore
	}
	if cfg.kill == nil {
		cfg.kill = syscall.Kill
	}
	if cfg.notify == nil {
		cfg.notify = signal.Notify
	}
	if cfg.stopNotify == nil {
		cfg.stopNotify = signal.Stop
	}
	return cfg
}

func prepareTerminal(cfg runtimeConfig, stdin *os.File, ptmx *os.File) (func() error, error) {
	stdinFD := int(stdin.Fd())
	if !cfg.isTerminal(stdinFD) {
		return nil, nil
	}

	if err := cfg.inheritSize(ptmx, stdin); err != nil {
		return nil, fmt.Errorf("inherit tty size: %w", err)
	}

	oldState, err := cfg.makeRaw(stdinFD)
	if err != nil {
		return nil, fmt.Errorf("set terminal raw mode: %w", err)
	}

	return func() error {
		if err := cfg.restore(stdinFD, oldState); err != nil {
			return fmt.Errorf("restore terminal state: %w", err)
		}
		return nil
	}, nil
}

func startSignalForwarder(
	cfg runtimeConfig,
	ptmx *os.File,
	stdin *os.File,
	childPID int,
) (chan struct{}, chan error, func()) {
	stop := make(chan struct{})
	done := make(chan struct{})
	errCh := make(chan error, 1)

	signals := cfg.signalSource
	var registered chan os.Signal
	if signals == nil {
		registered = make(chan os.Signal, 16)
		cfg.notify(
			registered,
			syscall.SIGWINCH,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
			syscall.SIGQUIT,
		)
		signals = registered
	}

	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				if registered != nil {
					cfg.stopNotify(registered)
				}
				return
			case sig := <-signals:
				if err := forwardSignal(cfg, sig, ptmx, stdin, childPID); err != nil {
					select {
					case errCh <- err:
					default:
					}
				}
			}
		}
	}()

	wait := func() {
		<-done
	}
	return stop, errCh, wait
}

func forwardSignal(
	cfg runtimeConfig,
	sig os.Signal,
	ptmx *os.File,
	stdin *os.File,
	childPID int,
) error {
	sysSig, ok := sig.(syscall.Signal)
	if !ok {
		return nil
	}

	switch sysSig {
	case syscall.SIGWINCH:
		if cfg.isTerminal(int(stdin.Fd())) {
			if err := cfg.inheritSize(ptmx, stdin); err != nil {
				return fmt.Errorf("forward SIGWINCH: %w", err)
			}
		}
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT:
		if err := cfg.kill(-childPID, sysSig); err != nil && !errors.Is(err, syscall.ESRCH) {
			return fmt.Errorf("forward %s to child session: %w", sysSig.String(), err)
		}
	}
	return nil
}

func asExitError(err error, target **exec.ExitError) bool {
	for err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			*target = exitErr
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}
