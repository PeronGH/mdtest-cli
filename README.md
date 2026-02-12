# mdtest

Run `.test.md` files using a local agent CLI (`claude` or `codex`).

## Requirements

- Go 1.25+
- `claude` or `codex` on `PATH`

## Run

```bash
go run ./cmd/mdtest run
```

For more options:

```bash
go run ./cmd/mdtest run -h
```

## Writing `.test.md` Files

`mdtest` prompts the agent with runtime details (test file path, output log path, and result-frontmatter contract).  
Your test file should focus on task behavior, not runner mechanics.

Use this structure:

```markdown
# Test Name

## Steps

1. Do concrete action A.
2. Verify concrete outcome B.
3. Set pass/fail decision based on observed result.
```

Guidelines:

- Write imperative, step-by-step instructions.
- Make assertions specific and observable.
- Avoid vague wording ("looks good", "works fine").
- Avoid meta instructions about where to write logs or YAML front matter.

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
