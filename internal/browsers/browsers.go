package browsers

import "os/exec"

// Browser is a detected browser and the command used to launch it.
type Browser struct {
	Name    string
	Command string
}

type candidate struct {
	name    string
	command string
}

// candidates lists the ten most popular desktop browsers and the PATH command
// names each is known by. This is the detection source on Linux and macOS; on
// Windows the registry (see detectPlatform) is authoritative instead. Entries
// are listed in popularity order, which sets the order of the results;
// duplicate resolved paths across aliases collapse to the first one found.
var candidates = []candidate{
	{"Google Chrome", "google-chrome"},
	{"Google Chrome", "google-chrome-stable"},
	{"Microsoft Edge", "microsoft-edge"},
	{"Microsoft Edge", "microsoft-edge-stable"},
	{"Firefox", "firefox"},
	{"Firefox", "firefox-esr"},
	{"Opera", "opera"},
	{"Brave", "brave-browser"},
	{"Brave", "brave-browser-stable"},
	{"Vivaldi", "vivaldi"},
	{"Vivaldi", "vivaldi-stable"},
	{"Chromium", "chromium"},
	{"Chromium", "chromium-browser"},
	{"Yandex Browser", "yandex-browser"},
	{"Yandex Browser", "yandex-browser-stable"},
	{"LibreWolf", "librewolf"},
	{"Tor Browser", "tor-browser"},
	{"Tor Browser", "torbrowser-launcher"},
}

// Detect returns the installed browsers. On Windows this reads the registry so
// browsers under Program Files (which are not on PATH) are found; there it is
// authoritative even when it finds nothing. Elsewhere it resolves the known
// command names on PATH.
func Detect() []Browser {
	if bs, handled := detectPlatform(); handled {
		return bs
	}
	return DetectWith(exec.LookPath)
}

// DetectWith is Detect with an injectable path lookup, for testing.
func DetectWith(lookPath func(string) (string, error)) []Browser {
	var out []Browser
	seen := map[string]bool{}
	for _, c := range candidates {
		full, err := lookPath(c.command)
		if err != nil || seen[full] {
			continue
		}
		seen[full] = true
		out = append(out, Browser{Name: c.name, Command: c.command})
	}
	return out
}
