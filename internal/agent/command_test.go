package agent

import (
	"reflect"
	"testing"
)

func TestCommandArgs(t *testing.T) {
	tests := []struct {
		name    string
		agent   Name
		prompt  string
		opts    CommandOptions
		want    []string
		wantErr bool
	}{
		{
			name:   "claude batch safe",
			agent:  ClaudeAgent,
			prompt: "p",
			want:   []string{"claude", "-p", "--permission-mode", "acceptEdits", "p"},
		},
		{
			name:   "claude batch dangerous",
			agent:  ClaudeAgent,
			prompt: "p",
			opts: CommandOptions{
				DangerouslyAllowAllActions: true,
			},
			want: []string{
				"claude",
				"-p",
				"--permission-mode", "acceptEdits",
				"--dangerously-skip-permissions",
				"p",
			},
		},
		{
			name:   "claude interactive safe",
			agent:  ClaudeAgent,
			prompt: "p",
			opts: CommandOptions{
				Interactive: true,
			},
			want: []string{
				"claude",
				"--permission-mode", "acceptEdits",
				"p",
			},
		},
		{
			name:   "claude interactive dangerous",
			agent:  ClaudeAgent,
			prompt: "p",
			opts: CommandOptions{
				Interactive:                true,
				DangerouslyAllowAllActions: true,
			},
			want: []string{
				"claude",
				"--permission-mode", "acceptEdits",
				"--dangerously-skip-permissions",
				"p",
			},
		},
		{
			name:   "codex batch safe",
			agent:  CodexAgent,
			prompt: "p",
			want:   []string{"codex", "exec", "p"},
		},
		{
			name:   "codex batch dangerous",
			agent:  CodexAgent,
			prompt: "p",
			opts: CommandOptions{
				DangerouslyAllowAllActions: true,
			},
			want: []string{
				"codex",
				"exec",
				"--dangerously-bypass-approvals-and-sandbox",
				"p",
			},
		},
		{
			name:   "codex interactive safe",
			agent:  CodexAgent,
			prompt: "p",
			opts: CommandOptions{
				Interactive: true,
			},
			want: []string{"codex", "p"},
		},
		{
			name:   "codex interactive dangerous",
			agent:  CodexAgent,
			prompt: "p",
			opts: CommandOptions{
				Interactive:                true,
				DangerouslyAllowAllActions: true,
			},
			want: []string{
				"codex",
				"--dangerously-bypass-approvals-and-sandbox",
				"p",
			},
		},
		{name: "invalid", agent: Name("other"), prompt: "p", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommandArgs(tt.agent, tt.prompt, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatal("CommandArgs returned nil error, want failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("CommandArgs returned error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CommandArgs = %#v, want %#v", got, tt.want)
			}
		})
	}
}
