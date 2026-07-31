package register

// RegEntry is a single registry value to write under HKEY_CURRENT_USER.
// An empty Name means the key's default value.
type RegEntry struct {
	Path  string
	Name  string
	Value string
}

const windowsAppKey = `Software\Clients\StartMenuInternet\LinkRouter`

// windowsRegEntries returns the registry values that register the app as a
// candidate default browser. It is pure so it can be tested on any OS.
func windowsRegEntries(execPath string) []RegEntry {
	capKey := windowsAppKey + `\Capabilities`
	cmd := `"` + execPath + `" open "%1"`
	return []RegEntry{
		{windowsAppKey, "", "Link Router"},
		{capKey, "ApplicationName", "Link Router"},
		{capKey, "ApplicationDescription", "Routes links to browsers by rule"},
		{capKey + `\URLAssociations`, "http", "LinkRouterHTTP"},
		{capKey + `\URLAssociations`, "https", "LinkRouterHTTP"},
		{windowsAppKey + `\shell\open\command`, "", cmd},
		{`Software\RegisteredApplications`, "LinkRouter", capKey},
		{`Software\Classes\LinkRouterHTTP\shell\open\command`, "", cmd},
	}
}
