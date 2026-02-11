# 001: `mdtest run` v1 Implementation Contract

Status: Accepted  
Date: 2026-02-11  
Design Reference: `docs/design/001-minimal-run-command.md`

## Purpose

Define the concrete implementation contract for the first runnable version of `mdtest run`.

This document is implementation-level: interfaces, invariants, and execution flow.

## Module Boundaries

1. `cmd/mdtest`: process entrypoint and top-level exit code handling.
2. `internal/cli`: command/flag wiring (`run` and `--agent`).
3. `internal/run`: suite orchestration.
4. `internal/agent`: agent mode resolution and command argv construction.
5. `internal/prompt`: prompt template rendering.
6. `internal/logs`: log path generation and front matter parsing.
7. `internal/ptyexec`: interactive child execution with terminal passthrough.

Equivalent package names are acceptable as long as these boundaries remain.

## Core Types

```go
type AgentMode string // "auto" | "claude" | "codex"
type AgentName string // "claude" | "codex"
type TestStatus string // "pass" | "fail"

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
    Reason  string // empty on success; diagnostic text on fail
}

type SuiteResult struct {
    Total  int
    Passed int
    Failed int
}
```

Field names may differ, but these values must be represented.

## Run Flow

`run` executes this fixed sequence:

1. Resolve root directory as absolute `pwd`.
2. Parse `--agent` flag (`auto|claude|codex`), reject other values.
3. Resolve effective agent once:
    - `auto`: prefer `claude`, fallback `codex`
    - explicit mode: require that binary on `PATH`
4. Discover tests recursively under root.
5. Sort tests by lexical relative path.
6. If no tests: return runner/setup error.
7. For each test, sequentially:
    - derive per-test log path
    - create log directory if missing
    - build prompt with absolute test path + absolute log path
    - run agent command in PTY passthrough
    - parse log front matter for `status`
    - record result (`pass`/`fail`)
8. Print suite summary.
9. Return exit code:
    - `0` all pass
    - `1` any fail
    - `2` runner/setup error

## Discovery Rules

1. Match files ending in `.test.md`.
2. Skip `.git/` directory subtree.
3. Ignore symlinked directories.
4. Use workspace-relative POSIX-style path strings in output and sorting.

## Log Path Rules

For `foo/bar/checkout.test.md`:

1. Base stem is `checkout` (remove `.test.md` suffix).
2. Log directory is sibling `foo/bar/checkout.logs/`.
3. Log filename is `<timestamp>.log.md`.
4. Timestamp is UTC `YYYY-MM-DDTHH-MM-SSZ`.

Collision rule:

1. If `<timestamp>.log.md` exists, append `-N` before `.log.md` (`-1`, `-2`, ...).

## Prompt Contract

Each test invocation must include these facts in prompt text:

1. Execute the test file step by step.
2. Read from exact absolute test path.
3. Write output to exact absolute log path.
4. Output must begin with YAML front matter containing `status: pass|fail`.

Prompt wording may evolve, but these required facts cannot be removed.

## Agent Command Construction

1. `claude`: argv `["claude", prompt]`
2. `codex`: argv `["codex", prompt]`
3. Working directory: suite root absolute path.
4. Child process exit code is recorded for diagnostics only.
5. Child process exit code never determines pass/fail.

## Log Front Matter Parser Contract

Parser input is one log file path. Output is `pass` or `fail`, else parse error.

Rules:

1. Only leading YAML front matter is parsed.
2. Front matter must start at byte 0 with `---` line.
3. Front matter must terminate with a closing `---` line.
4. Parsed YAML must include key `status`.
5. Accepted values are `pass` and `fail` (case-insensitive after trim).
6. Missing file, malformed YAML, missing key, or invalid value returns parse error.

Runner handling rule:

1. Any parse error marks that test failed and suite continues.
2. Parse errors are never fatal to the full run.

## PTY Passthrough Contract

`ptyexec` must guarantee:

1. Child process is attached to PTY.
2. Parent stdin/stdout are bridged to PTY for live interaction.
3. Parent terminal state is restored on normal exit, child error, and interrupts.
4. `SIGWINCH` and termination signals are forwarded to child session.

Terminal cleanup failure is a runner/setup error.

## Library Selection (Implementation-Level)

1. CLI/subcommands: `github.com/spf13/cobra`
2. PTY creation: `github.com/creack/pty`
3. Terminal raw-mode management: `golang.org/x/term`
4. YAML parsing for front matter: `gopkg.in/yaml.v3`

## Required Tests Before Code Completion

Go unit tests:

1. Agent resolution (`auto` preference and not-found errors).
2. Test discovery ordering and `.git`/symlink behavior.
3. Log path generation and collision suffix behavior.
4. Front matter parsing success/failure matrix.

Markdown integration tests (`.test.md`):

1. Time parity smoke test:
    - check current local time
    - if current minute is even, write `status: pass`
    - if current minute is odd, write `status: fail`
