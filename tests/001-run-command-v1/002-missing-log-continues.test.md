# Missing Log Marks Fail And Run Continues

Validate that when one test produces no log file, that test is marked failed and the suite still executes later tests.

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

cat > "$TMP_DIR/suite/a_missing.test.md" <<'EOF'
FAKE_BEHAVIOR=missing-log
EOF

cat > "$TMP_DIR/suite/b_pass.test.md" <<'EOF'
# inner pass
EOF

set +e
OUTPUT="$(cd "$TMP_DIR/suite" && PATH="$TMP_DIR/bin:$PATH" "$TMP_DIR/mdtest" run --agent codex 2>&1)"
CODE=$?
set -e

[ "$CODE" -eq 1 ]
printf '%s\n' "$OUTPUT" | grep -F "Total: 2, Passed: 1, Failed: 1"

PASS_LOG="$(find "$TMP_DIR/suite" -path '*/b_pass.logs/*.log.md' -type f | head -n1)"
[ -n "$PASS_LOG" ]
grep -F "status: pass" "$PASS_LOG"

if find "$TMP_DIR/suite" -path '*/a_missing.logs/*.log.md' -type f | grep -q .; then
	echo "expected no log file for a_missing.test.md"
	exit 1
fi
```
