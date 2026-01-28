package cli

import "testing"

func TestFmtCmd_HasChangedFlags(t *testing.T) {
	if fmtCmd.Flags().Lookup("changed") == nil {
		t.Fatal("fmtCmd should have --changed flag")
	}
	if fmtCmd.Flags().Lookup("ref") == nil {
		t.Fatal("fmtCmd should have --ref flag")
	}
}

func TestValCmd_HasChangedFlags(t *testing.T) {
	if valCmd.Flags().Lookup("changed") == nil {
		t.Fatal("valCmd should have --changed flag")
	}
	if valCmd.Flags().Lookup("ref") == nil {
		t.Fatal("valCmd should have --ref flag")
	}
}

func TestInitCmd_HasChangedFlags(t *testing.T) {
	if initCmd.Flags().Lookup("changed") == nil {
		t.Fatal("initCmd should have --changed flag")
	}
	if initCmd.Flags().Lookup("ref") == nil {
		t.Fatal("initCmd should have --ref flag")
	}
}

func TestPlanCmd_HasChangedFlags(t *testing.T) {
	if planCmd.Flags().Lookup("changed") == nil {
		t.Fatal("planCmd should have --changed flag")
	}
	if planCmd.Flags().Lookup("ref") == nil {
		t.Fatal("planCmd should have --ref flag")
	}
}

func TestTestCmd_HasChangedFlags(t *testing.T) {
	if testCmd.Flags().Lookup("changed") == nil {
		t.Fatal("testCmd should have --changed flag")
	}
	if testCmd.Flags().Lookup("ref") == nil {
		t.Fatal("testCmd should have --ref flag")
	}
}
