package cli

import (
	"testing"
)

func TestInitCmd_ExampleFlag(t *testing.T) {
	flag := initCmd.Flags().Lookup("example")
	if flag == nil {
		t.Fatal("initCmd should have --example flag")
	}

	if flag.Shorthand != "e" {
		t.Errorf("expected shorthand 'e', got '%s'", flag.Shorthand)
	}

	if flag.DefValue != "" {
		t.Errorf("expected empty default, got '%s'", flag.DefValue)
	}
}
