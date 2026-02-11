# Malformed Front Matter Marks Fail And Run Continues

Validate that malformed log front matter marks that test as failed while the suite continues.

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

cat > "$TMP_DIR/suite/a_bad.test.md" <<'EOF'
FAKE_BEHAVIOR=malformed
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

BAD_LOG="$(find "$TMP_DIR/suite" -path '*/a_bad.logs/*.log.md' -type f | head -n1)"
[ -n "$BAD_LOG" ]
head -n1 "$BAD_LOG" | grep -Fx "status: pass"

PASS_LOG="$(find "$TMP_DIR/suite" -path '*/b_pass.logs/*.log.md' -type f | head -n1)"
[ -n "$PASS_LOG" ]
grep -F "status: pass" "$PASS_LOG"
```
