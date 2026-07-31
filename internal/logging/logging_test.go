package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWriterLoggerFormats(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.Errorf("boom %d", 7)
	l.Infof("ok %s", "go")

	s := buf.String()
	if !strings.Contains(s, "ERROR: boom 7") {
		t.Errorf("missing error line in %q", s)
	}
	if !strings.Contains(s, "INFO: ok go") {
		t.Errorf("missing info line in %q", s)
	}
}

func TestNewFileLoggerAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "link-router.log")
	l, err := NewFileLogger(path)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}
	l.Errorf("first")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "ERROR: first") {
		t.Errorf("log file = %q, want it to contain %q", string(data), "ERROR: first")
	}
}
