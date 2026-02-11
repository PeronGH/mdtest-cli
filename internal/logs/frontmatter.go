package logs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"
)

func ParseStatus(path string) (Status, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read log: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	if !scanner.Scan() {
		return "", fmt.Errorf("missing front matter")
	}
	if strings.TrimSuffix(scanner.Text(), "\r") != "---" {
		return "", fmt.Errorf("front matter must start at byte 0 with ---")
	}

	yamlLines := make([]string, 0)
	closed := false
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "---" {
			closed = true
			break
		}
		yamlLines = append(yamlLines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan log: %w", err)
	}
	if !closed {
		return "", fmt.Errorf("missing closing front matter delimiter")
	}

	parsed := make(map[string]any)
	if err := yaml.Unmarshal([]byte(strings.Join(yamlLines, "\n")), &parsed); err != nil {
		return "", fmt.Errorf("parse yaml: %w", err)
	}

	raw, ok := parsed["status"]
	if !ok {
		return "", fmt.Errorf("missing status key")
	}

	normalized := strings.ToLower(strings.TrimSpace(fmt.Sprint(raw)))
	switch normalized {
	case string(StatusPass):
		return StatusPass, nil
	case string(StatusFail):
		return StatusFail, nil
	default:
		return "", fmt.Errorf("invalid status value %q", normalized)
	}
}
