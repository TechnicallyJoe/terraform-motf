package cli

import (
	"testing"
)

func TestValCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"example", "e"},
		{"init", "i"},
	}

	for _, tt := range tests {
		flag := valCmd.Flags().Lookup(tt.name)
		if flag == nil {
			t.Errorf("valCmd should have --%s flag", tt.name)
			continue
		}
		if flag.Shorthand != tt.shorthand {
			t.Errorf("--%s: expected shorthand '%s', got '%s'", tt.name, tt.shorthand, flag.Shorthand)
		}
	}
}
