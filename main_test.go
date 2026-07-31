package main

import (
	"bytes"
	"strings"
	"testing"
)

func newTestApp(args ...string) (*App, *bytes.Buffer, *bytes.Buffer) {
	var out, errb bytes.Buffer
	return &App{Args: args, Stdout: &out, Stderr: &errb}, &out, &errb
}

func TestExecuteNoArgsShowsUsage(t *testing.T) {
	app, _, errb := newTestApp()
	if code := app.Execute(); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), "usage") {
		t.Fatalf("stderr = %q, want it to contain %q", errb.String(), "usage")
	}
}

func TestExecuteHelp(t *testing.T) {
	app, out, _ := newTestApp("help")
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "usage") {
		t.Fatalf("stdout = %q, want it to contain %q", out.String(), "usage")
	}
}

func TestExecuteVersion(t *testing.T) {
	app, out, _ := newTestApp("version")
	if code := app.Execute(); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "link-router") {
		t.Fatalf("stdout = %q, want it to contain %q", out.String(), "link-router")
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	app, _, errb := newTestApp("frobnicate")
	if code := app.Execute(); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), "unknown command") {
		t.Fatalf("stderr = %q, want it to contain %q", errb.String(), "unknown command")
	}
}
