package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd_Output(t *testing.T) {
	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()

	// Should contain version info
	if !strings.Contains(output, "motf version") {
		t.Errorf("expected output to contain 'motf version', got: %s", output)
	}

	// Should contain commit info
	if !strings.Contains(output, "commit:") {
		t.Errorf("expected output to contain 'commit:', got: %s", output)
	}

	// Should contain build date info
	if !strings.Contains(output, "built:") {
		t.Errorf("expected output to contain 'built:', got: %s", output)
	}
}

func TestVersionCmd_MatchesFlag(t *testing.T) {
	// The version subcommand should produce the same output as --version flag
	expected := versionTemplate()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()

	if output != expected {
		t.Errorf("version command output doesn't match versionTemplate()\ngot:  %q\nwant: %q", output, expected)
	}
}

func TestVersionTemplate_Format(t *testing.T) {
	template := versionTemplate()

	// Should have three lines
	lines := strings.Split(strings.TrimSpace(template), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines in version template, got %d: %v", len(lines), lines)
	}

	// First line should start with "motf version"
	if !strings.HasPrefix(lines[0], "motf version") {
		t.Errorf("first line should start with 'motf version', got: %s", lines[0])
	}

	// Second line should start with "commit:"
	if !strings.HasPrefix(lines[1], "commit:") {
		t.Errorf("second line should start with 'commit:', got: %s", lines[1])
	}

	// Third line should start with "built:"
	if !strings.HasPrefix(lines[2], "built:") {
		t.Errorf("third line should start with 'built:', got: %s", lines[2])
	}
}
