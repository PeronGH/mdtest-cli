# 001: Minimal `mdtest run` Command

Status: Accepted  
Date: 2026-02-11

## Summary

First milestone is one command: `mdtest run`.

`mdtest run` discovers `*.test.md` files under `pwd`, delegates execution to an external agent CLI (`claude`/`codex`) with terminal passthrough, and decides pass/fail by parsing YAML front matter from agent-written log files.

This milestone does not parse front matter in test scripts. It only parses front matter in result logs.

## Scope

In scope:

1. Recursive discovery of `.test.md` files from `pwd`.
2. Agent selection (`auto`, `claude`, `codex`) and sequential execution.
3. PTY passthrough so agent CLI TUI interaction works normally.
4. Per-test log path generation and log front matter parsing for outcomes.
5. Suite summary and stable exit codes.

Out of scope:

1. Test-script front matter parsing (`requires`, `side-effects`).
2. Capability filtering and CI policy logic.
3. Parallel execution.
4. Windows support in v1.
5. Concrete Go library/package selection.

## User-Facing Behavior

### Command

```bash
mdtest run [--agent auto|claude|codex]
```

1. Default is `--agent auto`.
2. `auto` resolution order is `claude`, then `codex`.
3. If neither binary is on `PATH`, command exits with setup error.
4. `--agent claude` and `--agent codex` force explicit selection.

### Discovery

1. Walk `pwd` recursively.
2. Include paths ending in `.test.md`.
3. Skip `.git/`.
4. Ignore symlinked directories.
5. Run tests in lexical path order.
6. If no tests are found, exit with setup error.

### Per-Test Log Layout

For test file `X.test.md`, logs are written under sibling directory `X.logs/`.

Example:

1. `checkout.test.md`
2. `checkout.logs/2026-02-10T14-30-00Z.log.md`

Timestamp format is UTC RFC3339 with `:` replaced by `-`.

No screenshot directory is created by `mdtest` in this milestone.

### Prompt Contract

For each test, runner sends one prompt that includes:

1. Instruction to read and execute the test file step by step.
2. Exact absolute path of the test file.
3. Exact absolute path of the output log file to write.
4. Required log front matter contract:

```yaml
---
status: pass|fail
---
```

### Agent Invocation

Resolved agent is chosen once at run start.

1. `claude` mode executes `claude --permission-mode acceptEdits <PROMPT>`.
2. `codex` mode executes `codex <PROMPT>`.
3. Tests run sequentially in suite root directory.
4. Terminal interaction is passed through via PTY.

### Pass/Fail and Exit Codes

1. Process exit code from `claude`/`codex` is not used to decide test result.
2. Result is decided only by parsing `status` from the produced log front matter.
3. Missing log file, invalid front matter, or invalid/missing `status` counts as test `fail`.
4. Log parse/validation errors are non-fatal: mark current test `fail` and continue.
5. Runner continues through all tests and prints suite counts.
6. Exit `0` means all tests passed.
7. Exit `1` means one or more tests failed.
8. Exit `2` means runner/setup error (bad flags, missing agent binary, terminal/PTY failure, no tests found).

## Terminal Contract

1. Each child process runs under a PTY.
2. Parent terminal mode is restored on all exit paths.
3. Window resize and termination signals are forwarded.

## Acceptance Criteria

1. `mdtest run` discovers and executes `.test.md` files in deterministic order.
2. Default agent resolution is `claude` then `codex`.
3. Each test gets a log file in sibling `<basename>.logs/<timestamp>.log.md`.
4. Runner prompt includes absolute test path and absolute output log path.
5. Runner determines pass/fail only from parsed log front matter `status: pass|fail`.
6. Agent process exit code does not affect pass/fail.
7. Missing or invalid logs are recorded as failed tests.
8. Malformed log front matter never crashes the run; it marks the test failed.
9. Terminal state is restored after each test and on interruption.
10. No test-script front matter parsing exists in v1.
