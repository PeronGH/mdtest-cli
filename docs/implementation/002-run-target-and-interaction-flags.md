# 002: `mdtest run` Target and Interaction Flags Implementation Contract

Status: Accepted  
Date: 2026-02-12  
Design Reference: [`002`](../design/002-run-target-selection-and-execution-modes.md)  
Implementation Base: [`001`](001-run-command-v1.md)

## Intent

Define only implementation deltas from [`001`](001-run-command-v1.md). Unspecified behavior is unchanged.

## Module Deltas

1. `internal/cli`: add `--dir/-d`, positional file targets, `--interactive/-i`, `--dangerously-allow-all-actions/-A`.
2. `internal/run`: accept new config fields and run explicit file targets when provided.
3. `internal/agent`: build argv from agent + prompt + execution options.
4. `internal/procexec` (new): execute child process with or without PTY.

## Config and Request Deltas

```go
type Config struct {
    Root                        string
    Files                       []string
    Agent                       agent.Name
    Interactive                 bool
    DangerouslyAllowAllActions  bool
}

type ExecRequest struct {
    RootAbs     string
    Argv        []string
    Interactive bool
}
```

## CLI Contract

1. `--dir` default is `"."`.
2. Positional file arguments may be passed multiple times.
3. Effective root is always defined from `--dir` (explicit or default `"."`).
4. `--dir` and positional file arguments are valid together.
5. `--interactive` default is `false`.
6. `--dangerously-allow-all-actions` default is `false`.
7. `--agent` behavior is unchanged from [`001`](001-run-command-v1.md).

## Target Resolution Contract

When `len(cfg.Files) == 0`:

1. Use existing recursive discovery under `rootAbs`.

When `len(cfg.Files) > 0`:

1. Resolve `rootAbs` first from effective root.
2. For each positional file value:
   - if absolute: keep absolute path
   - if relative: join with `rootAbs`
   - clean and normalize path
   - require suffix `.test.md`
   - require regular file
   - require path to be inside `rootAbs`
3. Convert accepted paths to suite-relative POSIX form.
4. De-duplicate by relative path.
5. Sort lexically.
6. Empty final set is setup error.

## Agent Arg Construction Delta

Add execution options:

```go
type CommandOptions struct {
    Interactive                bool
    DangerouslyAllowAllActions bool
}
```

`CommandArgs(agent, prompt, opts)`:

1. `claude`:
   - batch: `claude -p --permission-mode acceptEdits <PROMPT>`
   - interactive: `claude --permission-mode acceptEdits <PROMPT>`
   - if dangerous: append `--dangerously-skip-permissions` before `<PROMPT>`
2. `codex`:
   - batch: `codex exec <PROMPT>`
   - interactive: `codex <PROMPT>`
   - if dangerous: append `--dangerously-bypass-approvals-and-sandbox` before `<PROMPT>`

Runner does not add other approval/sandbox flags by default.

## Process Execution Contract

`internal/procexec.Run(ctx, req)`:

1. `req.Interactive == true`: delegate to existing PTY executor behavior.
2. `req.Interactive == false`: run child without PTY using normal stdio inheritance.
3. Return child exit code; non-exit process errors remain runner/setup errors.

## Required Tests Before Code Completion

Go unit tests:

1. CLI parsing for new long and short flags plus positional file arguments.
2. Explicit target resolution matrix:
   - relative and absolute paths
   - outside-root rejection
   - missing file rejection
   - non-`.test.md` rejection
   - non-regular-file rejection
   - duplicate de-duplication
3. Agent argv matrix across:
   - `claude` vs `codex`
   - batch vs interactive
   - dangerous off vs on
4. Process executor dispatch: interactive uses PTY path, batch uses non-PTY path.

Manual verification commands (operator-run):

1. `mdtest run <path-a.test.md> <path-b.test.md>` runs exactly the selected files.
2. `mdtest run` runs non-interactive by default.
3. `mdtest run -i` runs with PTY passthrough.
