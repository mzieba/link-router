//go:build linux

package register

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const desktopFileName = "link-router.desktop"

func desktopFileContent(execPath string) string {
	return fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Link Router
Exec=%s open %%u
Terminal=false
Categories=Network;WebBrowser;
MimeType=x-scheme-handler/http;x-scheme-handler/https;
`, execPath)
}

func applicationsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "applications"), nil
}

func writeDesktopFile(dir, execPath string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, desktopFileName), []byte(desktopFileContent(execPath)), 0o644)
}

// Register installs the desktop entry and sets it as the default web browser.
func Register(execPath string) error {
	dir, err := applicationsDir()
	if err != nil {
		return err
	}
	if err := writeDesktopFile(dir, execPath); err != nil {
		return err
	}
	// Best-effort: these tools may be absent on minimal systems.
	_ = exec.Command("xdg-settings", "set", "default-web-browser", desktopFileName).Run()
	_ = exec.Command("xdg-mime", "default", desktopFileName, "x-scheme-handler/http", "x-scheme-handler/https").Run()
	return nil
}

// Unregister removes the desktop entry.
func Unregister() error {
	dir, err := applicationsDir()
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, desktopFileName)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
