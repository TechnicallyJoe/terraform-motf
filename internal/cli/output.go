package cli

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// ANSI color codes for terminal output
const (
	colorReset   = "\033[0m"
	colorCyan    = "\033[36m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorGreen   = "\033[32m"
	colorBlue    = "\033[34m"
	colorRed     = "\033[31m"
)

// colorPalette is the rotating list of colors for module prefixes
var colorPalette = []string{
	colorCyan,
	colorYellow,
	colorMagenta,
	colorGreen,
	colorBlue,
	colorRed,
}

// colorForIndex returns a color from the palette based on index
func colorForIndex(index int) string {
	return colorPalette[index%len(colorPalette)]
}

// prefixedWriter wraps an io.Writer and prepends a colored module prefix
// and timestamp to each line of output.
//
// Format: <color><module-name> |</color> HH:mm:ss.SSS # <message>
type prefixedWriter struct {
	out        io.Writer
	mu         *sync.Mutex
	buf        bytes.Buffer
	timeFunc   func() time.Time // for testing
	linePrefix string           // cached formatted prefix without timestamp
}

// newPrefixedWriter creates a new prefixedWriter.
// moduleName: the name of the module
// maxNameLen: maximum module name length for alignment
// colorIndex: index into color palette
// out: the underlying writer
// mu: mutex for thread-safe writing (shared across all writers)
func newPrefixedWriter(moduleName string, maxNameLen int, colorIndex int, out io.Writer, mu *sync.Mutex) *prefixedWriter {
	color := colorForIndex(colorIndex)
	// Pad the module name to align the | character
	paddedName := fmt.Sprintf("%-*s", maxNameLen, moduleName)
	linePrefix := fmt.Sprintf("%s%s |%s ", color, paddedName, colorReset)

	return &prefixedWriter{
		out:        out,
		linePrefix: linePrefix,
		mu:         mu,
		timeFunc:   time.Now,
	}
}

// Write implements io.Writer. It buffers input and writes complete lines
// with the prefix prepended.
func (w *prefixedWriter) Write(p []byte) (n int, err error) {
	written, err := w.buf.Write(p)
	if err != nil {
		return written, fmt.Errorf("failed to buffer output: %w", err)
	}

	for {
		line, readErr := w.buf.ReadBytes('\n')
		if readErr != nil {
			// No complete line yet, put it back
			if _, writeBackErr := w.buf.Write(line); writeBackErr != nil {
				return written, fmt.Errorf("failed to re-buffer partial line: %w", writeBackErr)
			}
			break
		}
		// Write the complete line with prefix
		if err := w.writeLine(line); err != nil {
			return len(p), err
		}
	}

	return len(p), nil
}

// Flush writes any remaining buffered content (for incomplete final lines)
func (w *prefixedWriter) Flush() error {
	if w.buf.Len() > 0 {
		line := w.buf.Bytes()
		w.buf.Reset()
		return w.writeLine(append(line, '\n'))
	}
	return nil
}

// writeLine writes a single line with prefix and timestamp
func (w *prefixedWriter) writeLine(line []byte) error {
	timestamp := w.timeFunc().Format("15:04:05.000")
	formatted := fmt.Sprintf("%s%s # %s", w.linePrefix, timestamp, string(line))

	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := w.out.Write([]byte(formatted))
	return err
}

// prefixedWriterPair holds stdout and stderr writers for a module
type prefixedWriterPair struct {
	stdout *prefixedWriter
	stderr *prefixedWriter
}

// newPrefixedWriterPair creates stdout and stderr writers for a module
func newPrefixedWriterPair(moduleName string, maxNameLen int, colorIndex int, stdout, stderr io.Writer, mu *sync.Mutex) *prefixedWriterPair {
	return &prefixedWriterPair{
		stdout: newPrefixedWriter(moduleName, maxNameLen, colorIndex, stdout, mu),
		stderr: newPrefixedWriter(moduleName, maxNameLen, colorIndex, stderr, mu),
	}
}

// Flush flushes both stdout and stderr writers
func (p *prefixedWriterPair) Flush() error {
	if err := p.stdout.Flush(); err != nil {
		return err
	}
	return p.stderr.Flush()
}
