# mdtest

Run `.test.md` files using a local agent CLI (`claude` or `codex`).

## Requirements

- Go 1.25+
- `claude` or `codex` on `PATH`

## Run

```bash
go run ./cmd/mdtest run
```

Or choose an agent explicitly:

```bash
go run ./cmd/mdtest run --agent auto
go run ./cmd/mdtest run --agent claude
go run ./cmd/mdtest run --agent codex
```

## Result Contract

Each test is passed to the agent. The agent writes a log file. `mdtest` reads status only from YAML front matter at byte 0:

```yaml
---
status: pass|fail
---
```

## Log Files

For `path/to/case.test.md`, logs are written to:
`path/to/case.logs/<timestamp>.log.md`

## Exit Codes

- `0`: all tests passed
- `1`: at least one test failed
- `2`: setup/runner error
