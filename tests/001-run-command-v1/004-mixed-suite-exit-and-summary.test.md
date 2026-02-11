# Mixed Suite Returns Exit 1 With Accurate Summary

Validate mixed outcomes: two passing tests and one failing test must yield `exit 1` and exact summary counts.

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

cat > "$TMP_DIR/suite/a_pass.test.md" <<'EOF'
# pass case
EOF

cat > "$TMP_DIR/suite/b_fail.test.md" <<'EOF'
FAKE_STATUS=fail
EOF

cat > "$TMP_DIR/suite/c_pass.test.md" <<'EOF'
# pass case
EOF

set +e
OUTPUT="$(cd "$TMP_DIR/suite" && PATH="$TMP_DIR/bin:$PATH" "$TMP_DIR/mdtest" run --agent codex 2>&1)"
CODE=$?
set -e

[ "$CODE" -eq 1 ]
printf '%s\n' "$OUTPUT" | grep -F "Total: 3, Passed: 2, Failed: 1"

FAIL_LOG="$(find "$TMP_DIR/suite" -path '*/b_fail.logs/*.log.md' -type f | head -n1)"
[ -n "$FAIL_LOG" ]
grep -F "status: fail" "$FAIL_LOG"
```
