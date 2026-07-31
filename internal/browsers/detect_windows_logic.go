package browsers

import (
	"path/filepath"
	"strings"
)

// winBrowser is one browser discovered in the registry: its display name and the
// raw shell\open\command string it launches with.
type winBrowser struct {
	name    string
	command string
}

// winRegReader abstracts the Windows registry reads used for browser detection.
// The real implementation lives in detect_windows.go; tests inject a fake so the
// detection logic can be exercised on any platform.
type winRegReader interface {
	// discover returns the browsers registered under StartMenuInternet, each
	// resolved to a display name and its raw shell\open\command string.
	discover() []winBrowser
	// appPath returns the App Paths full path registered for an exe basename
	// (for example "chrome.exe"), if present.
	appPath(exe string) (string, bool)
}

// detectWindows resolves installed browsers from the Windows registry. It takes
// the discovered browsers, resolves each to a full executable path (preferring
// the App Paths entry), verifies the file exists with exists, and deduplicates
// by path.
func detectWindows(reg winRegReader, exists func(string) bool) []Browser {
	var out []Browser
	seen := map[string]bool{}
	for _, b := range reg.discover() {
		exe := parseCommandExe(b.command)
		if exe == "" {
			continue
		}
		// Prefer the App Paths entry: a clean, canonical full path.
		if p, ok := reg.appPath(filepath.Base(exe)); ok && p != "" {
			exe = p
		}
		if !exists(exe) {
			continue
		}
		key := strings.ToLower(exe)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, Browser{Name: b.name, Command: exe})
	}
	return out
}

// parseCommandExe extracts the executable path from a Windows shell open command
// such as `"C:\Program Files\...\chrome.exe" -- "%1"` or
// `C:\path\firefox.exe -osint -url "%1"`.
func parseCommandExe(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return ""
	}
	if cmd[0] == '"' {
		if end := strings.IndexByte(cmd[1:], '"'); end >= 0 {
			return cmd[1 : 1+end]
		}
		return strings.TrimSpace(cmd[1:])
	}
	// Unquoted: take through the first ".exe" if present, otherwise the first
	// whitespace-delimited token.
	if i := strings.Index(strings.ToLower(cmd), ".exe"); i >= 0 {
		return cmd[:i+len(".exe")]
	}
	if sp := strings.IndexByte(cmd, ' '); sp >= 0 {
		return cmd[:sp]
	}
	return cmd
}
