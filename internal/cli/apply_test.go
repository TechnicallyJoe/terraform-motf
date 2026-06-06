package cli

import (
	"testing"
)

func TestApplyCmd_Flags(t *testing.T) {
	initFlagDef := applyCmd.Flags().Lookup("init")
	if initFlagDef == nil {
		t.Fatal("apply command should have --init flag")
	}
	if initFlagDef.Shorthand != "i" {
		t.Errorf("init flag shorthand = %q, want %q", initFlagDef.Shorthand, "i")
	}

	exampleFlagDef := applyCmd.Flags().Lookup("example")
	if exampleFlagDef == nil {
		t.Fatal("apply command should have --example flag")
	}
	if exampleFlagDef.Shorthand != "e" {
		t.Errorf("example flag shorthand = %q, want %q", exampleFlagDef.Shorthand, "e")
	}
}
