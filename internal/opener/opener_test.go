package opener

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mzieba/link-router/internal/config"
	"github.com/mzieba/link-router/internal/logging"
)

type call struct {
	command string
	args    []string
}

func cfg() *config.Config {
	return &config.Config{
		Default: "def",
		Rules:   []config.Rule{{MatchHost: []string{"*.work.com"}, Target: "work"}},
		Targets: map[string]config.Target{
			"def":   {Command: "firefox", Args: []string{"{url}"}},
			"work":  {Command: "chrome", Args: []string{"--profile-directory=Work", "{url}"}},
			"noarg": {Command: "opera", Args: []string{"--kiosk"}},
		},
	}
}

func TestOpenMatchedTargetSubstitutesURL(t *testing.T) {
	var got call
	run := func(cmd string, args []string) error { got = call{cmd, args}; return nil }
	if err := Open(cfg(), "https://a.work.com/x", run, logging.New(&bytes.Buffer{})); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got.command != "chrome" {
		t.Errorf("command = %q, want chrome", got.command)
	}
	want := []string{"--profile-directory=Work", "https://a.work.com/x"}
	if len(got.args) != 2 || got.args[1] != want[1] {
		t.Errorf("args = %v, want %v", got.args, want)
	}
}

func TestOpenNoMatchUsesDefault(t *testing.T) {
	var got call
	run := func(cmd string, args []string) error { got = call{cmd, args}; return nil }
	if err := Open(cfg(), "https://other.com", run, logging.New(&bytes.Buffer{})); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if got.command != "firefox" {
		t.Errorf("command = %q, want firefox", got.command)
	}
}

func TestOpenAppendsURLWhenNoToken(t *testing.T) {
	c := cfg()
	c.Default = "noarg"
	var got call
	run := func(cmd string, args []string) error { got = call{cmd, args}; return nil }
	if err := Open(c, "https://x.com", run, logging.New(&bytes.Buffer{})); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if len(got.args) != 2 || got.args[1] != "https://x.com" {
		t.Errorf("args = %v, want [--kiosk https://x.com]", got.args)
	}
}

func TestOpenFallsBackToDefaultOnFailure(t *testing.T) {
	var calls []call
	run := func(cmd string, args []string) error {
		calls = append(calls, call{cmd, args})
		if cmd == "chrome" {
			return errors.New("not found")
		}
		return nil
	}
	if err := Open(cfg(), "https://a.work.com", run, logging.New(&bytes.Buffer{})); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if len(calls) != 2 || calls[1].command != "firefox" {
		t.Fatalf("calls = %v, want chrome then firefox", calls)
	}
}

func TestOpenBothFail(t *testing.T) {
	run := func(cmd string, args []string) error { return errors.New("nope") }
	if err := Open(cfg(), "https://a.work.com", run, logging.New(&bytes.Buffer{})); err == nil {
		t.Fatal("want error when both targets fail")
	}
}

func TestResolveMatchedRule(t *testing.T) {
	res, err := Resolve(cfg(), "https://a.work.com/x")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !res.Matched {
		t.Error("Matched = false, want true")
	}
	if res.Target != "work" || res.Command != "chrome" {
		t.Errorf("target/command = %q/%q, want work/chrome", res.Target, res.Command)
	}
	want := []string{"--profile-directory=Work", "https://a.work.com/x"}
	if len(res.Args) != 2 || res.Args[1] != want[1] {
		t.Errorf("args = %v, want %v", res.Args, want)
	}
}

func TestResolveNoMatchUsesDefault(t *testing.T) {
	res, err := Resolve(cfg(), "https://other.com")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.Matched {
		t.Error("Matched = true, want false")
	}
	if res.Target != "def" || res.Command != "firefox" {
		t.Errorf("target/command = %q/%q, want def/firefox", res.Target, res.Command)
	}
}

func TestResolveUnknownTargetErrors(t *testing.T) {
	c := &config.Config{Default: "missing"}
	if _, err := Resolve(c, "https://x.com"); err == nil {
		t.Fatal("want error when the resolved target is not defined")
	}
}
