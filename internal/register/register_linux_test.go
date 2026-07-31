//go:build linux

package register

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDesktopFileContent(t *testing.T) {
	c := desktopFileContent("/usr/local/bin/link-router")
	for _, want := range []string{
		"MimeType=x-scheme-handler/http;x-scheme-handler/https;",
		"Exec=/usr/local/bin/link-router open %u",
		"Type=Application",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("desktop content missing %q:\n%s", want, c)
		}
	}
}

func TestWriteDesktopFile(t *testing.T) {
	dir := t.TempDir()
	if err := writeDesktopFile(dir, "/opt/link-router"); err != nil {
		t.Fatalf("writeDesktopFile: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, desktopFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "Exec=/opt/link-router open %u") {
		t.Errorf("written file missing Exec line:\n%s", string(data))
	}
}
