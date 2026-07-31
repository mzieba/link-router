//go:build windows

package register

import "golang.org/x/sys/windows/registry"

// Register writes the registry entries that make the app a default browser
// candidate. The user must still select it in Settings > Default apps.
func Register(execPath string) error {
	for _, e := range windowsRegEntries(execPath) {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, e.Path, registry.SET_VALUE)
		if err != nil {
			return err
		}
		err = k.SetStringValue(e.Name, e.Value)
		k.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Unregister removes the app's registry keys, best-effort.
func Unregister() error {
	subkeys := []string{
		windowsAppKey + `\shell\open\command`,
		windowsAppKey + `\shell\open`,
		windowsAppKey + `\shell`,
		windowsAppKey + `\Capabilities\URLAssociations`,
		windowsAppKey + `\Capabilities`,
		windowsAppKey,
		`Software\Classes\LinkRouterHTTP\shell\open\command`,
		`Software\Classes\LinkRouterHTTP\shell\open`,
		`Software\Classes\LinkRouterHTTP\shell`,
		`Software\Classes\LinkRouterHTTP`,
	}
	for _, s := range subkeys {
		_ = registry.DeleteKey(registry.CURRENT_USER, s)
	}
	if k, err := registry.OpenKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.SET_VALUE); err == nil {
		_ = k.DeleteValue("LinkRouter")
		k.Close()
	}
	return nil
}
