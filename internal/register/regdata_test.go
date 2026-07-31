package register

import (
	"strings"
	"testing"
)

func TestWindowsRegEntriesIncludeOpenCommand(t *testing.T) {
	entries := windowsRegEntries(`C:\Program Files\link-router\link-router.exe`)

	var foundCommand, foundRegistered bool
	for _, e := range entries {
		if strings.HasSuffix(e.Path, `\shell\open\command`) && e.Name == "" {
			if !strings.Contains(e.Value, "link-router.exe") || !strings.Contains(e.Value, `open "%1"`) {
				t.Errorf("open command value = %q, want it to launch open \"%%1\"", e.Value)
			}
			foundCommand = true
		}
		if e.Path == `Software\RegisteredApplications` && e.Name == "LinkRouter" {
			foundRegistered = true
		}
	}
	if !foundCommand {
		t.Error("no shell open command entry")
	}
	if !foundRegistered {
		t.Error("no RegisteredApplications entry")
	}
}
