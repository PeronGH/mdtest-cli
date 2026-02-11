# mdtest

`mdtest` runs Markdown test cases by delegating execution to an AI agent CLI.

Current milestone: `mdtest run` (v1).

## Requirements

- Go 1.25+
- At least one agent CLI on `PATH`:
  - `claude`
  - `codex`

## Quick Start

Run from your project root:

```bash
go run ./cmd/mdtest run
```

Or force an agent:

```bash
go run ./cmd/mdtest run --agent claude
go run ./cmd/mdtest run --agent codex
```

Agent modes:

- `auto` (default): prefer `claude`, then `codex`
- `claude`
- `codex`

## How It Works

1. Discovers `*.test.md` files recursively under the current directory.
2. Runs tests sequentially in lexical order.
3. For each test, asks the agent to read the test file and write a log file.
4. Decides pass/fail by parsing YAML front matter from that log:

```yaml
---
status: pass
---
```

Only `status: pass|fail` is used for outcome.

## Log Files

For a test like:

`tests/flow/checkout.test.md`

`mdtest` writes logs to:

`tests/flow/checkout.logs/<timestamp>.log.md`

If a timestamp collides, it appends `-1`, `-2`, etc.

## Exit Codes

- `0`: all tests passed
- `1`: at least one test failed
- `2`: setup/runner error
