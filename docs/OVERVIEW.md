# mdtest

Markdown as a testing language, with AI agents as the interpreter.

## Core Idea

Markdown is a scripting language for tests. An AI agent reads a Markdown file, executes the described steps, and reports whether the system under test behaves as expected. For the agent, this is manual execution. For the human, it's automatic.

With a computer-use agent, UI interactions are covered natively. Combined with MCP and Skills, tests can interact with real-world services — Cloudflare, Stripe, databases, anything the agent has access to.

## Test Authoring

A test is just a Markdown file. The simplest possible test has no front matter, no annotations — just natural language describing what to do and what to expect.

Tests that need specific runtime capabilities declare them in YAML front matter:

```markdown
---
requires: [browser]
---
```

```markdown
---
requires: [browser, mcp:cloudflare]
side-effects: true
---
```

If a test has no `requires` field, it runs everywhere. If a test has no `side-effects` field, it's assumed to have none.

### Side Effects

A single boolean. A test either produces side effects (modifies DNS, triggers payments, sends emails) or it doesn't. This distinction exists because side effects require human supervision — a developer at their machine can watch and intervene, but CI runs unsupervised.

- `side-effects: true` — the test modifies external state. Runs locally by default, excluded from CI unless the CI environment explicitly opts in.
- `side-effects: false` or absent — the test is pure from the outside world's perspective. Safe to run anywhere.

## Quality Control: The AI Linter

Precision in the test language is what ensures determinism. If the instructions are precise enough that a human tester could follow them without asking clarifying questions, an AI agent can execute them consistently.

An AI linter checks Markdown tests at authoring time for ambiguity and implicit assumptions. It flags:

- Vague assertions: "check that the page loads correctly" → specify the expected title, status code, or visible element.
- Implicit state dependencies: "click the third search result" without declaring what the search results should contain.
- Underspecified environments: steps that assume something about the system without declaring it in the front matter.

Well-linted tests have a side benefit: they serve as excellent documentation. Precise enough for deterministic agent execution means precise enough for a new team member to follow manually.

## Execution

### Runner

The runner reads the front matter of each test, compares `requires` against the current environment's declared capabilities, checks the `side-effects` policy, and decides what to run.

The logic is set intersection: if a test's requirements are a subset of the capabilities the runner detects, and the side-effects policy permits it, the test runs. The runner discovers available capabilities automatically — what MCP servers are connected, whether a browser agent is available, and so on. No environment configuration files needed.

### Developer Workflow

Developers are expected to run tests on their own machines before submitting a PR. Each developer has a Claude Max subscription as a standard part of their development environment — tokens are effectively free, so cost is not a constraint. The practical constraint is wall clock time.

- Change some code → run the fast, local, no-side-effect tests while still in flow.
- Adding a new feature → must include Markdown tests that run locally.
- Refactoring → run every test related to the changed code.

### CI Workflow

CI does not run every test on every commit. During PR CI, only a subset runs — specifically, tests that require no side effects and whose capabilities are available in the CI environment. Tests that were skipped are reported with the reason: "skipped because side effect policy does not permit" or "skipped because capability `mcp:cloudflare` is not available."

## Reporting: Execution Logs

The output of a test run is itself a Markdown file. The agent writes a narrative log of what it did, with YAML front matter summarizing the results.

```markdown
---
suite: checkout-flow
ran_at: 2026-02-10T14:30:00Z
passed: 3
failed: 1
cases:
    - name: add-to-cart
      status: pass
    - name: apply-coupon
      status: fail
      reason: "Expected discount of 20%, got 15%"
---

## Add to Cart ✅

Navigated to /products/widget-a. Clicked "Add to Cart."
Cart badge updated to show 1 item. Cart page shows Widget A at $29.99.

## Apply Coupon ❌

Entered coupon code `SAVE20` in the discount field. Clicked "Apply."
Banner displayed: "Coupon applied!"
**But**: Line total shows $25.49 (15% discount) instead of expected $23.99 (20% discount).

Screenshot: ![coupon-failure](./screenshots/apply-coupon-001.png)
```

The front matter is machine-parseable by any CI tool or dashboard. The body provides full investigative context in human-readable form. Since it's Markdown, it renders natively in GitHub comments, PRs, or any documentation tool.

Failed execution logs can be fed back to an agent for root cause analysis — the agent already has the full context of what was attempted and what happened.

## Incremental Conversion to Code

Markdown tests are the starting point, not necessarily the final form. The lifecycle is:

1. **Markdown-first**: Write tests in Markdown during the exploratory, rapidly-changing phase. Low cost to write, low cost to change, immediate coverage.
2. **Stabilization signal**: Over time, some features stabilize. The conversion advisor — an agent that analyzes git history, code churn, dependency graphs, and existing test coverage — identifies candidates for promotion.
3. **Selective promotion**: The advisor doesn't recommend converting entire Markdown tests wholesale. It maps which code paths each Markdown test exercises, checks what's already covered by coded tests, and suggests promoting only the portions that fill real gaps.
4. **Coded tests**: Important, stable paths get proper coded tests (Playwright, Cypress, pytest, etc.) for fast, deterministic CI execution.
5. **Permanent Markdown tier**: Some tests — especially complex multi-service orchestrations that would be painful to code — stay as Markdown permanently. Since tokens are free, the only reason to promote is execution speed and CI determinism.

### Conversion Signals

The advisor combines multiple weak signals into a strong recommendation:

- **Static analysis**: this module's interface hasn't changed in N commits.
- **Git history**: this feature area has had no PRs in M weeks.
- **Coverage analysis**: these code paths have no coded tests but are exercised by Markdown tests.
- **Dependency mapping**: this Markdown test touches services X, Y, Z — service X already has 90% unit coverage, service Z has nothing.

No single signal is sufficient. The combination, plus human judgment, determines when and what to convert.

## Architecture Summary

Four components, all operating on the same Markdown artifacts:

| Component              | Role                                                                                                        |
| ---------------------- | ----------------------------------------------------------------------------------------------------------- |
| **AI Linter**          | Checks test Markdown for ambiguity, vague assertions, and implicit state dependencies at authoring time.    |
| **Execution Agent**    | Reads Markdown tests, executes them via computer use and MCP, produces Markdown execution logs.             |
| **Reporting Layer**    | Structured YAML front matter for machine consumption, narrative Markdown body for human consumption.        |
| **Conversion Advisor** | Analyzes stability, coverage, and dependencies to recommend which Markdown tests to promote to coded tests. |

## Design Principles

- **Markdown all the way down.** Tests are Markdown. Execution logs are Markdown. Reports are Markdown. Everything is readable, diffable, and renderable without special tooling.
- **Declarative over descriptive.** Tests declare what they need (`requires`, `side-effects`), not where they should run. The environment decides what's permitted.
- **Zero-config default.** A plain Markdown file with no front matter is a valid test that runs everywhere with no side effects. Metadata is additive.
- **Start simple, promote when needed.** Markdown tests first, coded tests later, only for what's earned it. Complexity is introduced incrementally, never speculatively.
