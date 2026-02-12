package run

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverTests returns lexical, POSIX-style relative paths for *.test.md files.
func DiscoverTests(root string) ([]string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	tests := make([]string, 0)
	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if d.Type()&fs.ModeSymlink != 0 {
			isDir, err := symlinkPointsToDir(path)
			if err != nil {
				return err
			}
			if isDir {
				return filepath.SkipDir
			}
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".test.md") {
			return nil
		}

		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		tests = append(tests, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(tests)
	return tests, nil
}

// ResolveExplicitTests validates explicit file targets and returns lexical,
// POSIX-style relative paths inside rootAbs.
func ResolveExplicitTests(rootAbs string, files []string) ([]string, error) {
	unique := make(map[string]struct{}, len(files))

	for _, raw := range files {
		targetAbs := raw
		if !filepath.IsAbs(targetAbs) {
			targetAbs = filepath.Join(rootAbs, targetAbs)
		}
		targetAbs = filepath.Clean(targetAbs)

		rel, err := filepath.Rel(rootAbs, targetAbs)
		if err != nil {
			return nil, fmt.Errorf("resolve %q relative to root: %w", raw, err)
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
			return nil, fmt.Errorf("file %q is outside root %s", raw, rootAbs)
		}
		if !strings.HasSuffix(targetAbs, ".test.md") {
			return nil, fmt.Errorf("file %q must end with .test.md", raw)
		}

		info, err := os.Stat(targetAbs)
		if err != nil {
			return nil, fmt.Errorf("stat file %q: %w", raw, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("file %q is not a regular file", raw)
		}

		unique[filepath.ToSlash(rel)] = struct{}{}
	}

	tests := make([]string, 0, len(unique))
	for rel := range unique {
		tests = append(tests, rel)
	}
	sort.Strings(tests)
	return tests, nil
}

func symlinkPointsToDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}
