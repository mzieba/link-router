package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Logger records diagnostics from the open path.
type Logger interface {
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
}

// FileLogger writes plain-text lines to an io.Writer.
type FileLogger struct {
	w io.Writer
}

// New returns a logger that writes to w.
func New(w io.Writer) *FileLogger {
	return &FileLogger{w: w}
}

// NewFileLogger opens (appending, creating as needed) a log file at path.
func NewFileLogger(path string) (*FileLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileLogger{w: f}, nil
}

func (l *FileLogger) Errorf(format string, args ...any) {
	fmt.Fprintf(l.w, "ERROR: "+format+"\n", args...)
}

func (l *FileLogger) Infof(format string, args ...any) {
	fmt.Fprintf(l.w, "INFO: "+format+"\n", args...)
}
