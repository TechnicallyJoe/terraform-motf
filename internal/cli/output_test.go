package cli

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestColorForIndex(t *testing.T) {
	// Test that colors rotate through the palette
	tests := []struct {
		index    int
		expected string
	}{
		{0, colorCyan},
		{1, colorYellow},
		{2, colorMagenta},
		{3, colorGreen},
		{4, colorBlue},
		{5, colorRed},
		{6, colorCyan},   // wraps around
		{7, colorYellow}, // wraps around
	}

	for _, tt := range tests {
		got := colorForIndex(tt.index)
		if got != tt.expected {
			t.Errorf("colorForIndex(%d) = %q, want %q", tt.index, got, tt.expected)
		}
	}
}

func TestPrefixedWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}

	// Create a writer with a fixed time for testing
	fixedTime := time.Date(2025, 1, 31, 14, 32, 1, 123000000, time.UTC)
	pw := newPrefixedWriter("storage-account", 15, 0, &buf, mu)
	pw.timeFunc = func() time.Time { return fixedTime }

	// Write a complete line
	_, err := pw.Write([]byte("Hello, world!\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Check the format: <color>storage-account |</color> 14:32:01.123 # Hello, world!
	if !strings.Contains(output, "storage-account") {
		t.Errorf("output should contain module name, got: %s", output)
	}
	if !strings.Contains(output, "|") {
		t.Errorf("output should contain pipe separator, got: %s", output)
	}
	if !strings.Contains(output, "14:32:01.123") {
		t.Errorf("output should contain timestamp, got: %s", output)
	}
	if !strings.Contains(output, "# Hello, world!") {
		t.Errorf("output should contain message with # prefix, got: %s", output)
	}
	if !strings.Contains(output, colorCyan) {
		t.Errorf("output should contain cyan color code, got: %s", output)
	}
	if !strings.Contains(output, colorReset) {
		t.Errorf("output should contain color reset code, got: %s", output)
	}
}

func TestPrefixedWriter_WriteMultipleLines(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}

	fixedTime := time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC)
	pw := newPrefixedWriter("mod", 5, 0, &buf, mu)
	pw.timeFunc = func() time.Time { return fixedTime }

	// Write multiple lines at once
	_, err := pw.Write([]byte("Line 1\nLine 2\nLine 3\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}

	for i, line := range lines {
		if !strings.Contains(line, "mod") {
			t.Errorf("line %d should contain module name: %s", i, line)
		}
		if !strings.Contains(line, "#") {
			t.Errorf("line %d should contain # separator: %s", i, line)
		}
	}
}

func TestPrefixedWriter_WritePartialLines(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}

	fixedTime := time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC)
	pw := newPrefixedWriter("test", 10, 0, &buf, mu)
	pw.timeFunc = func() time.Time { return fixedTime }

	// Write partial content (no newline)
	_, err := pw.Write([]byte("partial"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("should not write until newline, got: %s", buf.String())
	}

	// Complete the line
	_, err = pw.Write([]byte(" content\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "partial content") {
		t.Errorf("output should contain complete message: %s", output)
	}
}

func TestPrefixedWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}

	fixedTime := time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC)
	pw := newPrefixedWriter("test", 10, 0, &buf, mu)
	pw.timeFunc = func() time.Time { return fixedTime }

	// Write partial content without newline
	_, err := pw.Write([]byte("incomplete"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("should not write until flush, got: %s", buf.String())
	}

	// Flush should write the remaining content
	err = pw.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "incomplete") {
		t.Errorf("flushed output should contain message: %s", output)
	}
}

func TestPrefixedWriter_Alignment(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}

	fixedTime := time.Date(2025, 1, 31, 10, 0, 0, 0, time.UTC)

	// Create two writers with different name lengths but same maxNameLen
	pw1 := newPrefixedWriter("short", 15, 0, &buf, mu)
	pw1.timeFunc = func() time.Time { return fixedTime }

	pw2 := newPrefixedWriter("very-long-name", 15, 1, &buf, mu)
	pw2.timeFunc = func() time.Time { return fixedTime }

	if _, err := pw1.Write([]byte("message 1\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if _, err := pw2.Write([]byte("message 2\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Find the position of | in each line (after stripping color codes)
	// The | should be at the same position due to padding
	stripColors := func(s string) string {
		// Simple color stripping for testing
		result := s
		for _, color := range []string{colorCyan, colorYellow, colorMagenta, colorGreen, colorBlue, colorRed, colorReset} {
			result = strings.ReplaceAll(result, color, "")
		}
		return result
	}

	pos1 := strings.Index(stripColors(lines[0]), "|")
	pos2 := strings.Index(stripColors(lines[1]), "|")

	if pos1 != pos2 {
		t.Errorf("pipe positions should be aligned: line1=%d, line2=%d\nLine1: %s\nLine2: %s",
			pos1, pos2, stripColors(lines[0]), stripColors(lines[1]))
	}
}

func TestPrefixedWriterPair(t *testing.T) {
	var stdout, stderr bytes.Buffer
	mu := &sync.Mutex{}

	pair := newPrefixedWriterPair("test-mod", 10, 0, &stdout, &stderr, mu)

	if _, err := pair.stdout.Write([]byte("stdout message\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if _, err := pair.stderr.Write([]byte("stderr message\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "stdout message") {
		t.Errorf("stdout should contain message: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "stderr message") {
		t.Errorf("stderr should contain message: %s", stderr.String())
	}
}

func TestPrefixedWriter_ConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}

	// Create multiple writers sharing the same mutex
	writers := make([]*prefixedWriter, 5)
	for i := range writers {
		writers[i] = newPrefixedWriter(fmt.Sprintf("mod-%d", i), 5, i, &buf, mu)
		writers[i].timeFunc = func() time.Time { return time.Now() }
	}

	// Write concurrently
	var wg sync.WaitGroup
	for i, w := range writers {
		wg.Add(1)
		go func(idx int, pw *prefixedWriter) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, _ = pw.Write([]byte("line\n"))
			}
		}(i, w)
	}
	wg.Wait()

	// Count lines - should have 50 lines (5 writers * 10 lines each)
	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 lines, got %d", len(lines))
	}

	// Each line should be well-formed (contain | and #)
	for i, line := range lines {
		if !strings.Contains(line, "|") || !strings.Contains(line, "#") {
			t.Errorf("line %d is malformed: %s", i, line)
		}
	}
}
