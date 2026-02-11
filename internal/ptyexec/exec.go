package ptyexec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Request struct {
	RootAbs string
	Argv    []string
}

type Result struct {
	ExitCode int
}

func Run(ctx context.Context, req Request) (Result, error) {
	if len(req.Argv) == 0 {
		return Result{}, fmt.Errorf("argv is empty")
	}

	cmd := exec.CommandContext(ctx, req.Argv[0], req.Argv[1:]...)
	cmd.Dir = req.RootAbs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := asExitError(err, &exitErr); ok {
			return Result{ExitCode: exitErr.ExitCode()}, nil
		}
		return Result{}, err
	}
	return Result{ExitCode: 0}, nil
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
