package cli

import (
	"testing"
)

// Tests for command flag registration

func TestInitCmd_HasExampleFlag(t *testing.T) {
	flag := initCmd.Flags().Lookup("example")
	if flag == nil {
		t.Fatal("initCmd should have --example flag")
	}
	if flag.Shorthand != "e" {
		t.Errorf("expected shorthand 'e', got '%s'", flag.Shorthand)
	}
}

func TestFmtCmd_HasExampleFlag(t *testing.T) {
	flag := fmtCmd.Flags().Lookup("example")
	if flag == nil {
		t.Fatal("fmtCmd should have --example flag")
	}
	if flag.Shorthand != "e" {
		t.Errorf("expected shorthand 'e', got '%s'", flag.Shorthand)
	}
}

func TestFmtCmd_HasInitFlag(t *testing.T) {
	flag := fmtCmd.Flags().Lookup("init")
	if flag == nil {
		t.Fatal("fmtCmd should have --init flag")
	}
	if flag.Shorthand != "i" {
		t.Errorf("expected shorthand 'i', got '%s'", flag.Shorthand)
	}
}

func TestValCmd_HasExampleFlag(t *testing.T) {
	flag := valCmd.Flags().Lookup("example")
	if flag == nil {
		t.Fatal("valCmd should have --example flag")
	}
	if flag.Shorthand != "e" {
		t.Errorf("expected shorthand 'e', got '%s'", flag.Shorthand)
	}
}

func TestValCmd_HasInitFlag(t *testing.T) {
	flag := valCmd.Flags().Lookup("init")
	if flag == nil {
		t.Fatal("valCmd should have --init flag")
	}
	if flag.Shorthand != "i" {
		t.Errorf("expected shorthand 'i', got '%s'", flag.Shorthand)
	}
}

func TestValCmd_HasValidateAlias(t *testing.T) {
	aliases := valCmd.Aliases
	found := false
	for _, alias := range aliases {
		if alias == "validate" {
			found = true
			break
		}
	}
	if !found {
		t.Error("valCmd should have 'validate' alias")
	}
}

func TestListCmd_HasSearchFlag(t *testing.T) {
	flag := listCmd.Flags().Lookup("search")
	if flag == nil {
		t.Fatal("listCmd should have --search flag")
	}
	if flag.Shorthand != "s" {
		t.Errorf("expected shorthand 's', got '%s'", flag.Shorthand)
	}
}

func TestRootCmd_HasPathFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("path")
	if flag == nil {
		t.Error("rootCmd should have --path persistent flag")
	}
}

func TestRootCmd_HasArgsFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("args")
	if flag == nil {
		t.Fatal("rootCmd should have --args persistent flag")
	}
	if flag.Shorthand != "a" {
		t.Errorf("expected shorthand 'a', got '%s'", flag.Shorthand)
	}
}

// Tests for command existence

func TestAllCommandsRegistered(t *testing.T) {
	commands := rootCmd.Commands()
	expectedCmds := []string{"init", "fmt", "val", "test", "list", "get", "config"}

	cmdMap := make(map[string]bool)
	for _, cmd := range commands {
		cmdMap[cmd.Name()] = true
	}

	for _, expected := range expectedCmds {
		if !cmdMap[expected] {
			t.Errorf("expected command '%s' to be registered on rootCmd", expected)
		}
	}
}
