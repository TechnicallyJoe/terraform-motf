package cli

import (
	"testing"
)

func TestTestCmd_NoExampleFlag(t *testing.T) {
	// test command doesn't support --example flag (tests run on module directly)
	flag := testCmd.Flags().Lookup("example")
	if flag != nil {
		t.Error("testCmd should not have --example flag")
	}
}
