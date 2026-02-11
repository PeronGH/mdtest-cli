package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseStatusSuccessCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Status
	}{
		{
			name:    "pass",
			content: "---\nstatus: pass\n---\nbody\n",
			want:    StatusPass,
		},
		{
			name:    "fail with trim and case",
			content: "---\nstatus: \"  FAIL  \"\n---\n",
			want:    StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeLog(t, tt.content)
			got, err := ParseStatus(path)
			if err != nil {
				t.Fatalf("ParseStatus returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseStatus = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStatusFailureCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		missing bool
	}{
		{name: "missing file", missing: true},
		{name: "front matter not at byte zero", content: "\n---\nstatus: pass\n---\n"},
		{name: "missing closing delimiter", content: "---\nstatus: pass\n"},
		{name: "malformed yaml", content: "---\nstatus: [\n---\n"},
		{name: "missing status key", content: "---\nother: pass\n---\n"},
		{name: "invalid status value", content: "---\nstatus: maybe\n---\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "case.log.md")
			if !tt.missing {
				if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}

			if _, err := ParseStatus(path); err == nil {
				t.Fatal("ParseStatus returned nil error, want failure")
			}
		})
	}
}

func writeLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "case.log.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
