package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNextLogPathBuildsSiblingLogDirAndTimestampName(t *testing.T) {
	root := t.TempDir()
	testAbs := filepath.Join(root, "foo", "checkout.test.md")
	if err := os.MkdirAll(filepath.Dir(testAbs), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	at := time.Date(2026, time.February, 10, 14, 30, 0, 0, time.UTC)
	logDir, logAbs, err := NextLogPath(testAbs, at)
	if err != nil {
		t.Fatalf("NextLogPath returned error: %v", err)
	}

	wantDir := filepath.Join(root, "foo", "checkout.logs")
	wantLog := filepath.Join(wantDir, "2026-02-10T14-30-00Z.log.md")

	if logDir != wantDir {
		t.Fatalf("logDir = %q, want %q", logDir, wantDir)
	}
	if logAbs != wantLog {
		t.Fatalf("logAbs = %q, want %q", logAbs, wantLog)
	}
}

func TestNextLogPathAddsCollisionSuffix(t *testing.T) {
	root := t.TempDir()
	testAbs := filepath.Join(root, "checkout.test.md")
	at := time.Date(2026, time.February, 10, 14, 30, 0, 0, time.UTC)

	logDir, first, err := NextLogPath(testAbs, at)
	if err != nil {
		t.Fatalf("NextLogPath returned error: %v", err)
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(first, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, second, err := NextLogPath(testAbs, at)
	if err != nil {
		t.Fatalf("NextLogPath returned error: %v", err)
	}
	if filepath.Base(second) != "2026-02-10T14-30-00Z-1.log.md" {
		t.Fatalf("second log filename = %q, want %q", filepath.Base(second), "2026-02-10T14-30-00Z-1.log.md")
	}
}
