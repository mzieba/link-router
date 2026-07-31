package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mzieba/link-router/internal/browsers"
	"github.com/mzieba/link-router/internal/config"
)

// detectVia adapts a fake PATH lookup into an App.Detect function.
func detectVia(lookPath func(string) (string, error)) func() []browsers.Browser {
	return func() []browsers.Browser { return browsers.DetectWith(lookPath) }
}

func appWith(t *testing.T, args ...string) (*App, *bytes.Buffer, string) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	var out bytes.Buffer
	app := &App{
		Args:       args,
		Stdout:     &out,
		Stderr:     &out,
		ConfigPath: cfgPath,
		Detect:     func() []browsers.Browser { return nil },
		ExecPath:   func() (string, error) { return "/opt/link-router", nil },
		Register:   func(string) error { return nil },
		Unregister: func() error { return nil },
	}
	return app, &out, cfgPath
}

func TestOpenCommandLaunchesMatchedTarget(t *testing.T) {
	app, _, cfgPath := appWith(t, "open", "https://a.work.com/x")
	if err := config.Save(cfgPath, &config.Config{
		Default: "def",
		Rules:   []config.Rule{{MatchHost: []string{"*.work.com"}, Target: "work"}},
		Targets: map[string]config.Target{
			"def":  {Command: "firefox", Args: []string{"{url}"}},
			"work": {Command: "chrome", Args: []string{"{url}"}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	var launched string
	app.Run = func(cmd string, args []string) error { launched = cmd; return nil }

	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if launched != "chrome" {
		t.Errorf("launched %q, want chrome", launched)
	}
}

func TestOpenCommandEmergencyFallbackWhenNoConfig(t *testing.T) {
	app, _, _ := appWith(t, "open", "https://x.com")
	app.Detect = detectVia(func(cmd string) (string, error) {
		if cmd == "firefox" {
			return "/usr/bin/firefox", nil
		}
		return "", errors.New("none")
	})
	var launched string
	app.Run = func(cmd string, args []string) error { launched = cmd; return nil }

	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if launched != "firefox" {
		t.Errorf("launched %q, want firefox (emergency)", launched)
	}
}

func TestBrowsersCommandLists(t *testing.T) {
	app, out, _ := appWith(t, "browsers")
	app.Detect = detectVia(func(cmd string) (string, error) {
		if cmd == "firefox" {
			return "/usr/bin/firefox", nil
		}
		return "", errors.New("none")
	})
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "firefox") {
		t.Errorf("output = %q, want it to list firefox", out.String())
	}
}

func TestOpenCommandEmergencyFailureIsLogged(t *testing.T) {
	app, _, cfgPath := appWith(t, "open", "https://x.com")
	// Detect finds nothing, so the emergency open has no browser to fall back to.
	app.Detect = func() []browsers.Browser { return nil }

	if code := app.Execute(); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}

	logPath := filepath.Join(filepath.Dir(cfgPath), "link-router.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "no config and no browser detected") {
		t.Errorf("log = %q, want it to contain the emergency-failure message", string(data))
	}
}

func TestBrowsersCommandDedupesByName(t *testing.T) {
	app, out, _ := appWith(t, "browsers")
	app.Detect = detectVia(func(cmd string) (string, error) {
		switch cmd {
		case "google-chrome":
			return "/usr/bin/google-chrome", nil
		case "google-chrome-stable":
			return "/usr/bin/google-chrome-stable", nil
		}
		return "", errors.New("none")
	})
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if n := strings.Count(out.String(), "Google Chrome"); n != 1 {
		t.Errorf("output contains %q %d times, want 1:\n%s", "Google Chrome", n, out.String())
	}
}

func TestListCommandShowsRulesAndDefault(t *testing.T) {
	app, out, cfgPath := appWith(t, "list")
	if err := config.Save(cfgPath, &config.Config{
		Default: "def",
		Rules: []config.Rule{
			{MatchHost: []string{"*.work.com"}, Target: "work"},
			{MatchURL: "^https://mail\\.", Target: "personal"},
		},
		Targets: map[string]config.Target{"def": {Command: "firefox", Args: []string{"{url}"}}},
	}); err != nil {
		t.Fatal(err)
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	got := out.String()
	for _, want := range []string{"Default target: def", "host *.work.com -> work", `url ^https://mail\. -> personal`} {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want it to contain %q", got, want)
		}
	}
}

func TestListCommandWithNoRules(t *testing.T) {
	app, out, cfgPath := appWith(t, "list")
	if err := config.Save(cfgPath, &config.Config{Default: "def"}); err != nil {
		t.Fatal(err)
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "No rules defined") {
		t.Errorf("output = %q, want it to report no rules", out.String())
	}
}

func TestTestCommandReportsMatchedTarget(t *testing.T) {
	app, out, cfgPath := appWith(t, "test", "https://a.work.com/x")
	if err := config.Save(cfgPath, &config.Config{
		Default: "def",
		Rules:   []config.Rule{{MatchHost: []string{"*.work.com"}, Target: "work"}},
		Targets: map[string]config.Target{
			"def":  {Command: "firefox", Args: []string{"{url}"}},
			"work": {Command: "chrome", Args: []string{"--profile-directory=Work", "{url}"}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, `target "work"`) {
		t.Errorf("output = %q, want it to name target work", got)
	}
	if !strings.Contains(got, "chrome --profile-directory=Work https://a.work.com/x") {
		t.Errorf("output = %q, want it to show the resolved command", got)
	}
}

func TestTestCommandReportsDefaultWhenNoRuleMatches(t *testing.T) {
	app, out, cfgPath := appWith(t, "test", "https://other.com")
	if err := config.Save(cfgPath, &config.Config{
		Default: "def",
		Targets: map[string]config.Target{"def": {Command: "firefox", Args: []string{"{url}"}}},
	}); err != nil {
		t.Fatal(err)
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "no rule") {
		t.Errorf("output = %q, want it to report no rule matched", out.String())
	}
}

func TestTestCommandRequiresURL(t *testing.T) {
	app, _, _ := appWith(t, "test")
	if code := app.Execute(); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestCommandsHintAtInitWhenConfigIsMissing(t *testing.T) {
	for _, args := range [][]string{{"test", "https://x.com"}, {"list"}} {
		t.Run(args[0], func(t *testing.T) {
			app, out, cfgPath := appWith(t, args...)
			if code := app.Execute(); code != 1 {
				t.Fatalf("exit code = %d, want 1", code)
			}
			got := out.String()
			if !strings.Contains(got, "link-router init") {
				t.Errorf("output = %q, want it to suggest running init", got)
			}
			if !strings.Contains(got, cfgPath) {
				t.Errorf("output = %q, want it to name the missing config path %q", got, cfgPath)
			}
		})
	}
}

func TestInitCommandCreatesConfig(t *testing.T) {
	app, out, cfgPath := appWith(t, "init")
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("init did not create config: %v", err)
	}
	if !strings.Contains(out.String(), "Created starter config") {
		t.Errorf("output = %q, want it to report the config was created", out.String())
	}
}

func TestInitCommandDoesNotOverwriteExistingConfig(t *testing.T) {
	app, out, cfgPath := appWith(t, "init")
	existing := &config.Config{
		Default: "keep",
		Targets: map[string]config.Target{"keep": {Command: "firefox", Args: []string{"{url}"}}},
	}
	if err := config.Save(cfgPath, existing); err != nil {
		t.Fatal(err)
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("output = %q, want it to report the config already exists", out.String())
	}
	c, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if c.Default != "keep" {
		t.Errorf("init overwrote existing config: default = %q, want %q", c.Default, "keep")
	}
}

func TestEditCommandOpensEditorWithConfigPath(t *testing.T) {
	app, _, cfgPath := appWith(t, "edit")
	app.Editor = func() string { return "vim" }
	var gotEditor, gotPath string
	app.EditFile = func(editor, path string) error {
		gotEditor, gotPath = editor, path
		return nil
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if gotEditor != "vim" || gotPath != cfgPath {
		t.Errorf("EditFile(%q, %q), want (%q, %q)", gotEditor, gotPath, "vim", cfgPath)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("edit did not create a starter config: %v", err)
	}
}

func TestEditCommandPrintsPathWithoutEditor(t *testing.T) {
	app, out, cfgPath := appWith(t, "edit")
	app.Editor = func() string { return "" }
	app.EditFile = func(string, string) error {
		t.Error("EditFile called even though no editor is configured")
		return nil
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), cfgPath) {
		t.Errorf("output = %q, want it to contain the config path %q", out.String(), cfgPath)
	}
}

func TestRegisterCommandCreatesConfig(t *testing.T) {
	app, _, cfgPath := appWith(t, "register")
	app.Run = func(string, []string) error { return nil }
	// Inject a fake Register so the test never touches the real OS.
	var registeredExec string
	registerCalled := false
	app.Register = func(execPath string) error {
		registerCalled = true
		registeredExec = execPath
		return nil
	}
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("register did not create config: %v", err)
	}
	if !registerCalled {
		t.Error("Register was not called")
	}
	if registeredExec != "/opt/link-router" {
		t.Errorf("Register called with %q, want %q", registeredExec, "/opt/link-router")
	}
}
