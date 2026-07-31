//go:build windows

package browsers

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// detectPlatform reads installed browsers from the Windows registry, so those
// under Program Files (which are not on PATH) are found and launched by full
// path. The second result is always true: on Windows the registry is the
// authoritative source, so Detect does not fall back to a PATH scan.
func detectPlatform() ([]Browser, bool) {
	return detectWindows(realReg{}, fileExists), true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type regLoc struct {
	root registry.Key
	base string
}

// StartMenuInternet lists registered browsers. It appears per-machine and
// per-user, plus a 32-bit view under WOW6432Node on 64-bit Windows.
var startMenuLocs = []regLoc{
	{registry.LOCAL_MACHINE, `SOFTWARE\Clients\StartMenuInternet`},
	{registry.CURRENT_USER, `SOFTWARE\Clients\StartMenuInternet`},
	{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Clients\StartMenuInternet`},
}

// App Paths maps an executable name to its full path.
var appPathLocs = []regLoc{
	{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths`},
	{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths`},
	{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\App Paths`},
}

// realReg implements winRegReader against the live Windows registry.
type realReg struct{}

func (realReg) discover() []winBrowser {
	var out []winBrowser
	seen := map[string]bool{}
	for _, loc := range startMenuLocs {
		k, err := registry.OpenKey(loc.root, loc.base, registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			continue
		}
		subkeys, err := k.ReadSubKeyNames(-1)
		k.Close()
		if err != nil {
			continue
		}
		for _, sub := range subkeys {
			key := strings.ToLower(sub)
			if seen[key] {
				continue
			}
			seen[key] = true

			// Read the command and friendly name from the same hive the
			// subkey was enumerated in.
			cmd, ok := readDefault(loc.root, loc.base+`\`+sub+`\shell\open\command`)
			if !ok {
				continue
			}
			name, _ := readDefault(loc.root, loc.base+`\`+sub)
			if name == "" {
				name = sub
			}
			out = append(out, winBrowser{name: name, command: cmd})
		}
	}
	return out
}

func (realReg) appPath(exe string) (string, bool) {
	for _, loc := range appPathLocs {
		if v, ok := readDefault(loc.root, loc.base+`\`+exe); ok && v != "" {
			return v, true
		}
	}
	return "", false
}

// readDefault returns the default ("") string value of the key at path.
func readDefault(root registry.Key, path string) (string, bool) {
	k, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return "", false
	}
	defer k.Close()
	v, _, err := k.GetStringValue("")
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(v), true
}
