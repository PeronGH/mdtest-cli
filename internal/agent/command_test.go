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
		want    []string
		wantErr bool
	}{
		{name: "claude", agent: ClaudeAgent, prompt: "p", want: []string{"claude", "p"}},
		{name: "codex", agent: CodexAgent, prompt: "p", want: []string{"codex", "p"}},
		{name: "invalid", agent: Name("other"), prompt: "p", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommandArgs(tt.agent, tt.prompt)
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
