package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NextLogPath returns the sibling log directory and next non-colliding log path.
func NextLogPath(testAbs string, at time.Time) (string, string, error) {
	base := filepath.Base(testAbs)
	if !strings.HasSuffix(base, ".test.md") {
		return "", "", fmt.Errorf("test path %q does not end with .test.md", testAbs)
	}

	stem := strings.TrimSuffix(base, ".test.md")
	logDir := filepath.Join(filepath.Dir(testAbs), stem+".logs")
	stamp := at.UTC().Format("2006-01-02T15-04-05Z")

	candidate := filepath.Join(logDir, stamp+".log.md")
	available, err := pathAvailable(candidate)
	if err != nil {
		return "", "", err
	}
	if available {
		return logDir, candidate, nil
	}

	for i := 1; ; i++ {
		candidate = filepath.Join(logDir, fmt.Sprintf("%s-%d.log.md", stamp, i))
		available, err := pathAvailable(candidate)
		if err != nil {
			return "", "", err
		}
		if available {
			return logDir, candidate, nil
		}
	}
}

func pathAvailable(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}
