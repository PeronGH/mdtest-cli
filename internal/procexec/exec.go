package procexec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/PeronGH/mdtest/internal/ptyexec"
)

type Request struct {
	RootAbs     string
	Argv        []string
	Interactive bool
}

type Result struct {
	ExitCode int
}

func Run(ctx context.Context, req Request) (Result, error) {
	return runWithDeps(ctx, req, runtimeDeps{
		runPTY:   runPTY,
		runBatch: runBatch,
	})
}

type runtimeDeps struct {
	runPTY   func(context.Context, Request) (Result, error)
	runBatch func(context.Context, Request) (Result, error)
}

func runWithDeps(ctx context.Context, req Request, deps runtimeDeps) (Result, error) {
	if len(req.Argv) == 0 {
		return Result{}, fmt.Errorf("argv is empty")
	}
	if deps.runPTY == nil {
		deps.runPTY = runPTY
	}
	if deps.runBatch == nil {
		deps.runBatch = runBatch
	}

	if req.Interactive {
		return deps.runPTY(ctx, req)
	}
	return deps.runBatch(ctx, req)
}

func runPTY(ctx context.Context, req Request) (Result, error) {
	res, err := ptyexec.Run(ctx, ptyexec.Request{
		RootAbs: req.RootAbs,
		Argv:    req.Argv,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{ExitCode: res.ExitCode}, nil
}

func runBatch(ctx context.Context, req Request) (Result, error) {
	cmd := exec.CommandContext(ctx, req.Argv[0], req.Argv[1:]...)
	cmd.Dir = req.RootAbs
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Result{ExitCode: exitErr.ExitCode()}, nil
		}
		return Result{}, err
	}
	return Result{ExitCode: 0}, nil
}
