package cmd

import (
	"testing"
)

func TestTaskCmd_Flags(t *testing.T) {
	taskFlagDef := taskCmd.Flags().Lookup("task")
	if taskFlagDef == nil {
		t.Error("task command should have --task flag")
	}
	if taskFlagDef.Shorthand != "t" {
		t.Errorf("task flag shorthand = %q, want %q", taskFlagDef.Shorthand, "t")
	}

	listFlagDef := taskCmd.Flags().Lookup("list")
	if listFlagDef == nil {
		t.Error("task command should have --list flag")
	}
	if listFlagDef.Shorthand != "l" {
		t.Errorf("list flag shorthand = %q, want %q", listFlagDef.Shorthand, "l")
	}
}
