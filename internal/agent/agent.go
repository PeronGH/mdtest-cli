package agent

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Mode string

const (
	AutoMode   Mode = "auto"
	ClaudeMode Mode = "claude"
	CodexMode  Mode = "codex"
)

type Name string

const (
	ClaudeAgent Name = "claude"
	CodexAgent  Name = "codex"
)

type LookPathFunc func(file string) (string, error)

type InvalidModeError struct {
	Raw string
}

func (e *InvalidModeError) Error() string {
	return fmt.Sprintf("invalid agent mode %q (expected auto, claude, or codex)", e.Raw)
}

type NotFoundError struct {
	Agent Name
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("agent %q was not found on PATH", e.Agent)
}

func ParseMode(raw string) (Mode, error) {
	mode := Mode(strings.TrimSpace(strings.ToLower(raw)))
	switch mode {
	case AutoMode, ClaudeMode, CodexMode:
		return mode, nil
	default:
		return "", &InvalidModeError{Raw: raw}
	}
}

func Resolve(mode Mode, lookPath LookPathFunc) (Name, error) {
	switch mode {
	case AutoMode:
		if exists(lookPath, string(ClaudeAgent)) {
			return ClaudeAgent, nil
		}
		if exists(lookPath, string(CodexAgent)) {
			return CodexAgent, nil
		}
		return "", &NotFoundError{Agent: ClaudeAgent}
	case ClaudeMode:
		return resolveExplicit(ClaudeAgent, lookPath)
	case CodexMode:
		return resolveExplicit(CodexAgent, lookPath)
	default:
		return "", &InvalidModeError{Raw: string(mode)}
	}
}

func exists(lookPath LookPathFunc, file string) bool {
	_, err := lookPath(file)
	return err == nil
}

func resolveExplicit(agent Name, lookPath LookPathFunc) (Name, error) {
	_, err := lookPath(string(agent))
	if err == nil {
		return agent, nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return "", &NotFoundError{Agent: agent}
	}
	return "", fmt.Errorf("resolve %q: %w", agent, err)
}

func CommandArgs(agent Name, prompt string) ([]string, error) {
	switch agent {
	case ClaudeAgent:
		return []string{string(ClaudeAgent), "--permission-mode", "acceptEdits", prompt}, nil
	case CodexAgent:
		return []string{string(CodexAgent), prompt}, nil
	default:
		return nil, fmt.Errorf("unsupported agent %q", agent)
	}
}
