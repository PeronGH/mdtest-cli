package run

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestResolveExplicitTestsAcceptsRelativeAndAbsoluteAndSorts(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "b.test.md"), "")
	mustWriteFile(t, filepath.Join(root, "a", "one.test.md"), "")

	got, err := ResolveExplicitTests(root, []string{
		filepath.Join(root, "b.test.md"),
		"a/one.test.md",
	})
	if err != nil {
		t.Fatalf("ResolveExplicitTests returned error: %v", err)
	}

	want := []string{"a/one.test.md", "b.test.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveExplicitTests = %#v, want %#v", got, want)
	}
}

func TestResolveExplicitTestsDeduplicatesNormalizedPaths(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "a", "one.test.md"), "")

	got, err := ResolveExplicitTests(root, []string{
		"a/one.test.md",
		"./a/./one.test.md",
		filepath.Join(root, "a", "one.test.md"),
	})
	if err != nil {
		t.Fatalf("ResolveExplicitTests returned error: %v", err)
	}

	want := []string{"a/one.test.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveExplicitTests = %#v, want %#v", got, want)
	}
}

func TestResolveExplicitTestsRejectsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	mustWriteFile(t, filepath.Join(outside, "outside.test.md"), "")

	_, err := ResolveExplicitTests(root, []string{filepath.Join(outside, "outside.test.md")})
	if err == nil {
		t.Fatal("ResolveExplicitTests returned nil error, want outside-root rejection")
	}
	if !strings.Contains(err.Error(), "outside root") {
		t.Fatalf("ResolveExplicitTests error = %q, want outside-root text", err.Error())
	}
}

func TestResolveExplicitTestsRejectsMissingFile(t *testing.T) {
	root := t.TempDir()
	_, err := ResolveExplicitTests(root, []string{"missing.test.md"})
	if err == nil {
		t.Fatal("ResolveExplicitTests returned nil error, want missing file rejection")
	}
}

func TestResolveExplicitTestsRejectsNonTestMarkdownSuffix(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "note.md"), "")

	_, err := ResolveExplicitTests(root, []string{"note.md"})
	if err == nil {
		t.Fatal("ResolveExplicitTests returned nil error, want .test.md suffix rejection")
	}
	if !strings.Contains(err.Error(), ".test.md") {
		t.Fatalf("ResolveExplicitTests error = %q, want .test.md text", err.Error())
	}
}

func TestResolveExplicitTestsRejectsNonRegularFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "dir.test.md"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	_, err := ResolveExplicitTests(root, []string{"dir.test.md"})
	if err == nil {
		t.Fatal("ResolveExplicitTests returned nil error, want non-regular rejection")
	}
	if !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("ResolveExplicitTests error = %q, want regular-file text", err.Error())
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
