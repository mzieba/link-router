package browsers

import (
	"errors"
	"testing"
)

func TestDetectWithFindsPresentSkipsMissing(t *testing.T) {
	present := map[string]string{
		"firefox":       "/usr/bin/firefox",
		"google-chrome": "/usr/bin/google-chrome",
	}
	lookPath := func(cmd string) (string, error) {
		if p, ok := present[cmd]; ok {
			return p, nil
		}
		return "", errors.New("not found")
	}

	got := DetectWith(lookPath)
	if len(got) != 2 {
		t.Fatalf("detected %d browsers, want 2: %v", len(got), got)
	}
	found := map[string]bool{}
	for _, b := range got {
		found[b.Command] = true
	}
	if !found["firefox"] || !found["google-chrome"] {
		t.Errorf("detected %v, want both firefox and google-chrome", got)
	}
}

func TestDetectWithDedupesByResolvedPath(t *testing.T) {
	lookPath := func(cmd string) (string, error) {
		if cmd == "google-chrome" || cmd == "google-chrome-stable" {
			return "/usr/bin/google-chrome", nil
		}
		return "", errors.New("not found")
	}
	got := DetectWith(lookPath)
	if len(got) != 1 {
		t.Fatalf("detected %d, want 1 (deduped): %v", len(got), got)
	}
}

func TestDetectWithFindsPopularBrowsers(t *testing.T) {
	// Each of the ten popular browsers should be found by at least one of its
	// known command names.
	byCommand := map[string]string{
		"google-chrome":  "Google Chrome",
		"microsoft-edge": "Microsoft Edge",
		"firefox":        "Firefox",
		"chromium":       "Chromium",
		"brave-browser":  "Brave",
		"opera":          "Opera",
		"vivaldi":        "Vivaldi",
		"yandex-browser": "Yandex Browser",
		"librewolf":      "LibreWolf",
		"tor-browser":    "Tor Browser",
	}
	lookPath := func(cmd string) (string, error) {
		if _, ok := byCommand[cmd]; ok {
			return "/usr/bin/" + cmd, nil
		}
		return "", errors.New("not found")
	}

	got := DetectWith(lookPath)
	foundNames := map[string]bool{}
	for _, b := range got {
		foundNames[b.Name] = true
	}
	for _, name := range byCommand {
		if !foundNames[name] {
			t.Errorf("did not detect %q; detected: %v", name, got)
		}
	}
}

func TestParseCommandExe(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`"C:\Program Files\Google\Chrome\Application\chrome.exe" -- "%1"`, `C:\Program Files\Google\Chrome\Application\chrome.exe`},
		{`C:\Program Files\Mozilla Firefox\firefox.exe -osint -url "%1"`, `C:\Program Files\Mozilla Firefox\firefox.exe`},
		{`"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"`, `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`},
		{`  "C:\x\brave.exe"  --incognito`, `C:\x\brave.exe`},
		{`chrome.exe`, `chrome.exe`},
		{``, ``},
	}
	for _, c := range cases {
		if got := parseCommandExe(c.in); got != c.want {
			t.Errorf("parseCommandExe(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// fakeReg is an in-memory winRegReader for testing detectWindows off-Windows.
type fakeReg struct {
	browsers []winBrowser
	appPaths map[string]string
}

func (f fakeReg) discover() []winBrowser { return f.browsers }
func (f fakeReg) appPath(exe string) (string, bool) {
	p, ok := f.appPaths[exe]
	return p, ok
}

func TestDetectWindowsResolvesFullPaths(t *testing.T) {
	reg := fakeReg{
		browsers: []winBrowser{
			{name: "Google Chrome", command: `"C:\Program Files\Google\Chrome\Application\chrome.exe" -- "%1"`},
			{name: "Mozilla Firefox", command: `"C:\Program Files\Mozilla Firefox\firefox.exe" -osint -url "%1"`},
		},
		// App Paths gives a clean full path for chrome even though the command matched.
		appPaths: map[string]string{
			"chrome.exe": `C:\Program Files\Google\Chrome\Application\chrome.exe`,
		},
	}
	existsAll := func(string) bool { return true }

	got := detectWindows(reg, existsAll)
	if len(got) != 2 {
		t.Fatalf("detected %d browsers, want 2: %v", len(got), got)
	}
	if got[0].Name != "Google Chrome" {
		t.Errorf("first name = %q, want Google Chrome", got[0].Name)
	}
	if got[0].Command != `C:\Program Files\Google\Chrome\Application\chrome.exe` {
		t.Errorf("first command = %q, want the full chrome.exe path", got[0].Command)
	}
	if got[1].Name != "Mozilla Firefox" {
		t.Errorf("second name = %q, want Mozilla Firefox", got[1].Name)
	}
	if got[1].Command != `C:\Program Files\Mozilla Firefox\firefox.exe` {
		t.Errorf("second command = %q, want the full firefox.exe path", got[1].Command)
	}
}

func TestDetectWindowsSkipsMissingExecutables(t *testing.T) {
	reg := fakeReg{
		browsers: []winBrowser{
			{name: "Ghost Browser", command: `"C:\gone\ghost.exe" "%1"`},
		},
	}
	// The registry lists it, but the exe no longer exists on disk.
	got := detectWindows(reg, func(string) bool { return false })
	if len(got) != 0 {
		t.Fatalf("detected %d browsers, want 0 (exe missing): %v", len(got), got)
	}
}

func TestDetectWindowsDedupesByPath(t *testing.T) {
	// HKLM and HKCU can both register the same browser under different subkeys.
	reg := fakeReg{
		browsers: []winBrowser{
			{name: "Google Chrome", command: `"C:\Program Files\Google\Chrome\Application\chrome.exe" "%1"`},
			{name: "Google Chrome", command: `"C:\Program Files\Google\Chrome\Application\CHROME.EXE" "%1"`},
		},
	}
	got := detectWindows(reg, func(string) bool { return true })
	if len(got) != 1 {
		t.Fatalf("detected %d browsers, want 1 (deduped): %v", len(got), got)
	}
}
