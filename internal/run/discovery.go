package run

import (
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

func symlinkPointsToDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}
