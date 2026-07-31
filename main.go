package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/mzieba/link-router/internal/browsers"
	"github.com/mzieba/link-router/internal/config"
	"github.com/mzieba/link-router/internal/opener"
	"github.com/mzieba/link-router/internal/register"
)

// it's overridden at build time via -ldflags "-X main.version=<value>" (see Makefile).
var version = "dev"

// appVersion resolves the version to display. It prefers the ldflags-injected
// value, then falls back to the module version Go embeds for
// `go install ...@vX.Y.Z` builds, and finally to the "dev" default.
func appVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

type App struct {
	Args       []string
	Stdout     io.Writer
	Stderr     io.Writer
	ConfigPath string
	Run        opener.Runner
	Detect     func() []browsers.Browser
	ExecPath   func() (string, error)
	Register   func(execPath string) error
	Unregister func() error
	Editor     func() string
	EditFile   func(editor, path string) error
}

// defaultEditor returns the user's preferred editor command, or "" if neither
// $VISUAL nor $EDITOR is set.
func defaultEditor() string {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

// runEditor runs editor against path attached to the terminal and waits for it
// to exit. editor may carry its own flags, e.g. "code --wait".
func runEditor(editor, path string) error {
	fields := strings.Fields(editor)
	cmd := exec.Command(fields[0], append(fields[1:], path)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func main() {
	cfgPath, _ := config.DefaultPath()
	app := &App{
		Args:       os.Args[1:],
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		ConfigPath: cfgPath,
		Run:        opener.DefaultRunner,
		Detect:     browsers.Detect,
		ExecPath:   os.Executable,
		Register:   register.Register,
		Unregister: register.Unregister,
		Editor:     defaultEditor,
		EditFile:   runEditor,
	}
	os.Exit(app.Execute())
}

func (a *App) Execute() int {
	if len(a.Args) == 0 {
		a.usage(a.Stderr)
		return 2
	}
	cmd, rest := a.Args[0], a.Args[1:]
	var err error
	switch cmd {
	case "open":
		err = a.cmdOpen(rest)
	case "list":
		err = a.cmdList()
	case "test":
		err = a.cmdTest(rest)
	case "init":
		err = a.cmdInit()
	case "register":
		err = a.cmdRegister()
	case "unregister":
		err = a.cmdUnregister()
	case "browsers":
		err = a.cmdBrowsers()
	case "edit":
		err = a.cmdEdit()
	case "version":
		err = a.cmdVersion()
	case "-h", "--help", "help":
		a.usage(a.Stdout)
		return 0
	default:
		fmt.Fprintf(a.Stderr, "unknown command %q\n", cmd)
		a.usage(a.Stderr)
		return 2
	}
	if err != nil {
		fmt.Fprintln(a.Stderr, "error:", err)
		return 1
	}
	return 0
}

func (a *App) usage(w io.Writer) {
	fmt.Fprint(w, `usage: link-router <command> [args]

commands:
  open <url>    open a URL using the matching rule
  list          list the configured rules and default target
  test <url>    show which rule and target a URL resolves to
  init          create a starter config file if none exists
  register      register as a default browser candidate
  unregister    remove the default browser registration
  browsers      list detected browsers
  edit          open the config in $VISUAL or $EDITOR
  version       print the application version
`)
}
