package cli

import (
	"bytes"
	"runtime/debug"
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

// mockBuildInfo returns a function that can be assigned to readBuildInfo for testing
func mockBuildInfo(info *debug.BuildInfo, ok bool) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return info, ok
	}
}

func TestEffectiveVersion_LdflagsWin(t *testing.T) {
	// Save and restore globals
	origVersion, origCommit, origDate := version, commit, date
	origReadBuildInfo := readBuildInfo
	defer func() {
		version, commit, date = origVersion, origCommit, origDate
		readBuildInfo = origReadBuildInfo
	}()

	// Set ldflags values (non-defaults)
	version = "v1.2.3"
	commit = "abc123"
	date = "2026-01-01T00:00:00Z"

	// Mock build info with different values
	readBuildInfo = mockBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Version: "v9.9.9"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "zzz999"},
			{Key: "vcs.time", Value: "2099-12-31T23:59:59Z"},
		},
	}, true)

	v, c, d := effectiveVersion()

	// ldflags should win
	if v != "v1.2.3" {
		t.Errorf("version: got %q, want %q", v, "v1.2.3")
	}
	if c != "abc123" {
		t.Errorf("commit: got %q, want %q", c, "abc123")
	}
	if d != "2026-01-01T00:00:00Z" {
		t.Errorf("date: got %q, want %q", d, "2026-01-01T00:00:00Z")
	}
}

func TestEffectiveVersion_FallbackToModuleVersion(t *testing.T) {
	origVersion, origCommit, origDate := version, commit, date
	origReadBuildInfo := readBuildInfo
	defer func() {
		version, commit, date = origVersion, origCommit, origDate
		readBuildInfo = origReadBuildInfo
	}()

	// Defaults (ldflags not set)
	version, commit, date = "dev", "none", "unknown"

	// Mock build info with module version only
	readBuildInfo = mockBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Version: "v0.4.0"},
	}, true)

	v, c, d := effectiveVersion()

	if v != "v0.4.0" {
		t.Errorf("version: got %q, want %q", v, "v0.4.0")
	}
	// commit/date stay default when VCS info not available
	if c != "none" {
		t.Errorf("commit: got %q, want %q", c, "none")
	}
	if d != "unknown" {
		t.Errorf("date: got %q, want %q", d, "unknown")
	}
}

func TestEffectiveVersion_FallbackToVCSInfo(t *testing.T) {
	origVersion, origCommit, origDate := version, commit, date
	origReadBuildInfo := readBuildInfo
	defer func() {
		version, commit, date = origVersion, origCommit, origDate
		readBuildInfo = origReadBuildInfo
	}()

	version, commit, date = "dev", "none", "unknown"

	readBuildInfo = mockBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Version: "v0.5.0"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "deadbeef1234"},
			{Key: "vcs.time", Value: "2026-01-26T10:00:00Z"},
		},
	}, true)

	v, c, d := effectiveVersion()

	if v != "v0.5.0" {
		t.Errorf("version: got %q, want %q", v, "v0.5.0")
	}
	if c != "deadbeef1234" {
		t.Errorf("commit: got %q, want %q", c, "deadbeef1234")
	}
	if d != "2026-01-26T10:00:00Z" {
		t.Errorf("date: got %q, want %q", d, "2026-01-26T10:00:00Z")
	}
}

func TestEffectiveVersion_ModifiedSuffix(t *testing.T) {
	origVersion, origCommit, origDate := version, commit, date
	origReadBuildInfo := readBuildInfo
	defer func() {
		version, commit, date = origVersion, origCommit, origDate
		readBuildInfo = origReadBuildInfo
	}()

	version, commit, date = "dev", "none", "unknown"

	readBuildInfo = mockBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Version: "v0.6.0"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abc123"},
			{Key: "vcs.modified", Value: "true"},
		},
	}, true)

	v, c, _ := effectiveVersion()

	if v != "v0.6.0" {
		t.Errorf("version: got %q, want %q", v, "v0.6.0")
	}
	if c != "abc123 (modified)" {
		t.Errorf("commit: got %q, want %q", c, "abc123 (modified)")
	}
}

func TestEffectiveVersion_DevelVersionKeptAsDev(t *testing.T) {
	origVersion, origCommit, origDate := version, commit, date
	origReadBuildInfo := readBuildInfo
	defer func() {
		version, commit, date = origVersion, origCommit, origDate
		readBuildInfo = origReadBuildInfo
	}()

	version, commit, date = "dev", "none", "unknown"

	// Local build: Main.Version is "(devel)"
	readBuildInfo = mockBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "localcommit"},
			{Key: "vcs.time", Value: "2026-01-26T12:00:00Z"},
		},
	}, true)

	v, c, d := effectiveVersion()

	// Version stays "dev" because (devel) is not useful
	if v != "dev" {
		t.Errorf("version: got %q, want %q", v, "dev")
	}
	// But commit/date can still be filled from VCS
	if c != "localcommit" {
		t.Errorf("commit: got %q, want %q", c, "localcommit")
	}
	if d != "2026-01-26T12:00:00Z" {
		t.Errorf("date: got %q, want %q", d, "2026-01-26T12:00:00Z")
	}
}

func TestEffectiveVersion_NoBuildInfo(t *testing.T) {
	origVersion, origCommit, origDate := version, commit, date
	origReadBuildInfo := readBuildInfo
	defer func() {
		version, commit, date = origVersion, origCommit, origDate
		readBuildInfo = origReadBuildInfo
	}()

	version, commit, date = "dev", "none", "unknown"

	// Build info not available
	readBuildInfo = mockBuildInfo(nil, false)

	v, c, d := effectiveVersion()

	// All defaults remain
	if v != "dev" {
		t.Errorf("version: got %q, want %q", v, "dev")
	}
	if c != "none" {
		t.Errorf("commit: got %q, want %q", c, "none")
	}
	if d != "unknown" {
		t.Errorf("date: got %q, want %q", d, "unknown")
	}
}
