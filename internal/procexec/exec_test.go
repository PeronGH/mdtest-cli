package procexec

import (
	"context"
	"errors"
	"testing"
)

func TestRunWithDepsDispatchesPTYWhenInteractive(t *testing.T) {
	ptyCalled := false
	batchCalled := false

	got, err := runWithDeps(context.Background(), Request{
		RootAbs:     t.TempDir(),
		Argv:        []string{"echo", "hello"},
		Interactive: true,
	}, runtimeDeps{
		runPTY: func(context.Context, Request) (Result, error) {
			ptyCalled = true
			return Result{ExitCode: 11}, nil
		},
		runBatch: func(context.Context, Request) (Result, error) {
			batchCalled = true
			return Result{}, nil
		},
	})
	if err != nil {
		t.Fatalf("runWithDeps returned error: %v", err)
	}
	if got.ExitCode != 11 {
		t.Fatalf("ExitCode = %d, want 11", got.ExitCode)
	}
	if !ptyCalled {
		t.Fatal("PTY runner was not called")
	}
	if batchCalled {
		t.Fatal("batch runner was called unexpectedly")
	}
}

func TestRunWithDepsDispatchesBatchWhenNonInteractive(t *testing.T) {
	ptyCalled := false
	batchCalled := false

	got, err := runWithDeps(context.Background(), Request{
		RootAbs: t.TempDir(),
		Argv:    []string{"echo", "hello"},
	}, runtimeDeps{
		runPTY: func(context.Context, Request) (Result, error) {
			ptyCalled = true
			return Result{}, nil
		},
		runBatch: func(context.Context, Request) (Result, error) {
			batchCalled = true
			return Result{ExitCode: 7}, nil
		},
	})
	if err != nil {
		t.Fatalf("runWithDeps returned error: %v", err)
	}
	if got.ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", got.ExitCode)
	}
	if ptyCalled {
		t.Fatal("PTY runner was called unexpectedly")
	}
	if !batchCalled {
		t.Fatal("batch runner was not called")
	}
}

func TestRunBatchReturnsChildExitCode(t *testing.T) {
	got, err := runBatch(context.Background(), Request{
		RootAbs: t.TempDir(),
		Argv:    []string{"sh", "-c", "exit 9"},
	})
	if err != nil {
		t.Fatalf("runBatch returned error: %v", err)
	}
	if got.ExitCode != 9 {
		t.Fatalf("ExitCode = %d, want 9", got.ExitCode)
	}
}

func TestRunBatchReturnsErrorForProcessStartFailure(t *testing.T) {
	_, err := runBatch(context.Background(), Request{
		RootAbs: t.TempDir(),
		Argv:    []string{"definitely-not-a-real-command"},
	})
	if err == nil {
		t.Fatal("runBatch returned nil error, want failure")
	}
}

func TestRunWithDepsRejectsEmptyArgv(t *testing.T) {
	_, err := runWithDeps(context.Background(), Request{RootAbs: t.TempDir()}, runtimeDeps{
		runPTY: func(context.Context, Request) (Result, error) { return Result{}, nil },
		runBatch: func(context.Context, Request) (Result, error) {
			return Result{}, errors.New("should not be called")
		},
	})
	if err == nil {
		t.Fatal("runWithDeps returned nil error, want failure")
	}
}
