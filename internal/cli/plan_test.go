package cli

import (
	"testing"
)

func TestPlanCmd_Flags(t *testing.T) {
	// Verify init flag is registered
	initFlagDef := planCmd.Flags().Lookup("init")
	if initFlagDef == nil {
		t.Fatal("plan command should have --init flag")
	}
	if initFlagDef.Shorthand != "i" {
		t.Errorf("init flag shorthand = %q, want %q", initFlagDef.Shorthand, "i")
	}

	// Verify example flag is registered
	exampleFlagDef := planCmd.Flags().Lookup("example")
	if exampleFlagDef == nil {
		t.Fatal("plan command should have --example flag")
	}
	if exampleFlagDef.Shorthand != "e" {
		t.Errorf("example flag shorthand = %q, want %q", exampleFlagDef.Shorthand, "e")
	}
}
