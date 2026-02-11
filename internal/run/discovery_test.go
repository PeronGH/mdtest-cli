package run

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverTestsSkipsDotGitAndSymlinkDirsAndSortsLexical(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "b.test.md"), "")
	mustWriteFile(t, filepath.Join(root, "a", "one.test.md"), "")
	mustWriteFile(t, filepath.Join(root, "a", "two.md"), "")
	mustWriteFile(t, filepath.Join(root, ".git", "ignored.test.md"), "")

	external := t.TempDir()
	mustWriteFile(t, filepath.Join(external, "symlinked.test.md"), "")
	if err := os.Symlink(external, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlink unsupported in this environment: %v", err)
	}

	got, err := DiscoverTests(root)
	if err != nil {
		t.Fatalf("DiscoverTests returned error: %v", err)
	}

	want := []string{"a/one.test.md", "b.test.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DiscoverTests = %#v, want %#v", got, want)
	}
}

func TestDiscoverTestsReturnsPosixRelativePaths(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "nested", "case.test.md"), "")

	got, err := DiscoverTests(root)
	if err != nil {
		t.Fatalf("DiscoverTests returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("DiscoverTests returned %d items, want 1", len(got))
	}
	if got[0] != "nested/case.test.md" {
		t.Fatalf("DiscoverTests path = %q, want %q", got[0], "nested/case.test.md")
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
