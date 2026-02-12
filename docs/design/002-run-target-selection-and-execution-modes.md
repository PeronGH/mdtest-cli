# 002: `mdtest run` Target and Interaction Flags

Status: Accepted  
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

1. Effective root is `--dir` if provided, otherwise implicit `.`.
2. `--dir` sets suite root and child process working directory.
3. `--dir` and `--file` can be used together.
4. Without `--file`, runner discovers `*.test.md` recursively under the effective root.
5. With one or more `--file`, discovery is skipped and runner executes exactly the provided set.
6. For `--file` values:
   - relative paths are resolved from the effective root
   - absolute paths are allowed only if they are inside the effective root
   - each path must exist, be a regular file, and end with `.test.md`
7. Duplicate targets are removed after normalization.
8. Final run order is lexical by suite-relative path.
9. Empty target set is a setup error.

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
3. `mdtest run -d tests -f smoke/a.test.md` resolves `-f` from `-d` and runs only that target.
4. Invalid `--file` values return setup error with clear diagnostics.
5. `mdtest run` runs non-interactive by default.
6. `mdtest run -i` runs with PTY passthrough.
7. `-A` toggles the correct dangerous flag per agent.
