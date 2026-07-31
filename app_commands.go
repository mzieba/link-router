package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mzieba/link-router/internal/config"
	"github.com/mzieba/link-router/internal/logging"
	"github.com/mzieba/link-router/internal/opener"
)

func (a *App) cmdOpen(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: link-router open <url>")
	}
	rawURL := args[0]
	log := a.logger()

	c, err := config.Load(a.ConfigPath)
	if err != nil {
		log.Errorf("load config: %v", err)
		return a.emergencyOpen(rawURL, log)
	}
	return opener.Open(c, rawURL, a.Run, log)
}

// loadConfig loads the config for commands that require one, replacing the bare
// "no such file or directory" for a missing config with a pointer to `init`.
func (a *App) loadConfig() (*config.Config, error) {
	c, err := config.Load(a.ConfigPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("no config file at %s\nhint: run 'link-router init' to create a starter config", a.ConfigPath)
	}
	return c, err
}

func (a *App) emergencyOpen(rawURL string, log logging.Logger) error {
	bs := a.Detect()
	if len(bs) == 0 {
		err := fmt.Errorf("no config and no browser detected for %q", rawURL)
		log.Errorf("%v", err)
		return err
	}
	log.Infof("no usable config; using detected browser %q", bs[0].Command)
	return a.Run(bs[0].Command, []string{rawURL})
}

func (a *App) cmdBrowsers() error {
	bs := a.Detect()
	if len(bs) == 0 {
		fmt.Fprintln(a.Stdout, "No browsers detected.")
		return nil
	}
	seen := make(map[string]bool, len(bs))
	for _, b := range bs {
		if seen[b.Name] {
			continue
		}
		seen[b.Name] = true
		fmt.Fprintf(a.Stdout, "%s\t%s\n", b.Name, b.Command)
	}
	return nil
}

func (a *App) cmdRegister() error {
	exe, err := a.ExecPath()
	if err != nil {
		return err
	}
	if _, err := a.ensureConfig(); err != nil {
		return err
	}
	if err := a.Register(exe); err != nil {
		return err
	}
	fmt.Fprintln(a.Stdout, "Registered Link Router.")
	fmt.Fprintln(a.Stdout, "On Windows, open Settings > Default apps and select Link Router to finish.")
	return nil
}

func (a *App) cmdUnregister() error {
	if err := a.Unregister(); err != nil {
		return err
	}
	fmt.Fprintln(a.Stdout, "Unregistered Link Router.")
	return nil
}

func (a *App) cmdList() error {
	c, err := a.loadConfig()
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "Default target: %s\n", c.Default)
	if len(c.Rules) == 0 {
		fmt.Fprintln(a.Stdout, "No rules defined; every URL uses the default target.")
		return nil
	}
	fmt.Fprintln(a.Stdout, "Rules (evaluated top to bottom):")
	for i, r := range c.Rules {
		var matchers []string
		if len(r.MatchHost) > 0 {
			matchers = append(matchers, "host "+strings.Join(r.MatchHost, ", "))
		}
		if r.MatchURL != "" {
			matchers = append(matchers, "url "+r.MatchURL)
		}
		if len(matchers) == 0 {
			matchers = append(matchers, "(no matchers; skipped)")
		}
		fmt.Fprintf(a.Stdout, "  %d. %s -> %s\n", i+1, strings.Join(matchers, "; "), r.Target)
	}
	return nil
}

func (a *App) cmdTest(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: link-router test <url>")
	}
	rawURL := args[0]

	c, err := a.loadConfig()
	if err != nil {
		return err
	}
	res, err := opener.Resolve(c, rawURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "URL:     %s\n", rawURL)
	if res.Matched {
		fmt.Fprintf(a.Stdout, "Matched: rule -> target %q\n", res.Target)
	} else {
		fmt.Fprintf(a.Stdout, "Matched: no rule; using default target %q\n", res.Target)
	}
	fmt.Fprintf(a.Stdout, "Command: %s\n", strings.Join(append([]string{res.Command}, res.Args...), " "))
	return nil
}

func (a *App) cmdInit() error {
	created, err := a.ensureConfig()
	if err != nil {
		return err
	}
	if !created {
		fmt.Fprintln(a.Stdout, "Config already exists at", a.ConfigPath)
		return nil
	}
	fmt.Fprintln(a.Stdout, "Created starter config at", a.ConfigPath)
	fmt.Fprintln(a.Stdout, "Edit it to set your browsers and rules.")
	return nil
}

// cmdEdit opens the config in $VISUAL or $EDITOR, creating a starter config
// first if none exists. With neither variable set it prints the path instead.
func (a *App) cmdEdit() error {
	if _, err := a.ensureConfig(); err != nil {
		return err
	}
	var editor string
	if a.Editor != nil {
		editor = a.Editor()
	}
	if editor == "" {
		fmt.Fprintln(a.Stdout, "No $VISUAL or $EDITOR set. Edit the config file directly:")
		fmt.Fprintln(a.Stdout, a.ConfigPath)
		return nil
	}
	return a.EditFile(editor, a.ConfigPath)
}

func (a *App) cmdVersion() error {
	fmt.Fprintf(a.Stdout, "link-router %s\n", appVersion())
	return nil
}

// ensureConfig creates a starter config if none exists, reporting whether it
// created the file.
func (a *App) ensureConfig() (created bool, err error) {
	if _, err := os.Stat(a.ConfigPath); err == nil {
		return false, nil
	}
	if err := config.Save(a.ConfigPath, config.Sample()); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) logger() logging.Logger {
	l, err := logging.NewFileLogger(filepath.Join(filepath.Dir(a.ConfigPath), "link-router.log"))
	if err != nil {
		return logging.New(a.Stderr)
	}
	return l
}
