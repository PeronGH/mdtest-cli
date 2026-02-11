#!/usr/bin/env sh
set -eu

prompt="${1:-}"
test_path="$(printf '%s\n' "$prompt" | sed -n 's#^Read the test from this exact absolute path: ##p' | head -n1)"
log_path="$(printf '%s\n' "$prompt" | sed -n 's#^Write the output log to this exact absolute path: ##p' | head -n1)"

if [ -z "$test_path" ] || [ -z "$log_path" ]; then
	exit 3
fi

if grep -q '^FAKE_BEHAVIOR=missing-log$' "$test_path"; then
	exit 0
fi

mkdir -p "$(dirname "$log_path")"

if grep -q '^FAKE_BEHAVIOR=malformed$' "$test_path"; then
	cat >"$log_path" <<'EOF'
status: pass
EOF
	exit 0
fi

status="pass"
if grep -q '^FAKE_STATUS=fail$' "$test_path"; then
	status="fail"
fi

cat >"$log_path" <<EOF
---
status: $status
---
fake-codex
EOF
