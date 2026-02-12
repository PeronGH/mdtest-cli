# 002: `mdtest run` Target and Interaction Flags

Status: Proposed  
Date: 2026-02-12  
Design Base: [`001`](001-minimal-run-command.md)

## Intent

Define only the behavior changes from [`001`](001-minimal-run-command.md). Any behavior not listed here is unchanged.

## CLI Delta

```bash
mdtest run \
  [--agent|-a auto|claude|codex] \
  [--dir|-d <path>] \
  [--file|-f <path> ...] \
  [--interactive|-i] \
  [--dangerously-allow-all-actions|-A]
```

Defaults:

1. `--agent auto`
2. `--dir .`
3. `--file` omitted
4. `--interactive` off
5. `--dangerously-allow-all-actions` off

## Target Selection

1. `--dir` sets suite root and child process working directory.
2. Without `--file`, runner discovers `*.test.md` recursively under `--dir`.
3. With `--file`, runner executes exactly the provided set.
4. Every `--file` must exist, be a regular file, end with `.test.md`, and resolve within `--dir`.
5. Duplicate targets are removed after normalization.
6. Final run order is lexical by suite-relative path.
7. Empty target set is a setup error.

## Execution Style

Default (`--interactive` off):

1. Run without PTY.
2. Agent commands:
    - `claude`: `claude -p --permission-mode acceptEdits <PROMPT>`
    - `codex`: `codex exec <PROMPT>`

Interactive (`--interactive` on):

1. Run with PTY passthrough.
2. Agent commands:
    - `claude`: `claude --permission-mode acceptEdits <PROMPT>`
    - `codex`: `codex <PROMPT>`

## Dangerous Mode

When `--dangerously-allow-all-actions` is set, append:

1. `claude`: `--dangerously-skip-permissions`
2. `codex`: `--dangerously-bypass-approvals-and-sandbox`

This mode is explicit opt-in only.

## Acceptance

1. `mdtest run -d tests/smoke` runs discovery under `tests/smoke`.
2. `mdtest run -f tests/a.test.md -f tests/b.test.md` runs exactly those files.
3. Invalid `--file` values return setup error with clear diagnostics.
4. `mdtest run` runs non-interactive by default.
5. `mdtest run -i` runs with PTY passthrough.
6. `-A` toggles the correct dangerous flag per agent.
