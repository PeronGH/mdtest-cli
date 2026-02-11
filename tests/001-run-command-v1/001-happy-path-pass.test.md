# `mdtest run` Happy Path (`status: pass`)

Validate that a valid log front matter with `status: pass` yields suite success (`exit 0`).

## Steps

```bash
set -euo pipefail

REPO_ROOT="$(pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

go build -o "$TMP_DIR/mdtest" ./cmd/mdtest

mkdir -p "$TMP_DIR/bin" "$TMP_DIR/suite"
cp "$REPO_ROOT/tests/helpers/fake-codex.sh" "$TMP_DIR/bin/codex"
chmod +x "$TMP_DIR/bin/codex"

cat > "$TMP_DIR/suite/happy.test.md" <<'EOF'
# inner test
This inner test should pass.
EOF

set +e
OUTPUT="$(cd "$TMP_DIR/suite" && PATH="$TMP_DIR/bin:$PATH" "$TMP_DIR/mdtest" run --agent codex 2>&1)"
CODE=$?
set -e

[ "$CODE" -eq 0 ]
printf '%s\n' "$OUTPUT" | grep -F "Total: 1, Passed: 1, Failed: 0"

LOG_FILE="$(find "$TMP_DIR/suite" -path '*/happy.logs/*.log.md' -type f | head -n1)"
[ -n "$LOG_FILE" ]
grep -F "status: pass" "$LOG_FILE"
```
