package config

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	want := Sample()
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Default != want.Default {
		t.Errorf("Default = %q, want %q", got.Default, want.Default)
	}
	if len(got.Rules) != len(want.Rules) {
		t.Fatalf("len(Rules) = %d, want %d", len(got.Rules), len(want.Rules))
	}
	if got.Targets["work"].Command != want.Targets["work"].Command {
		t.Errorf("work command = %q, want %q", got.Targets["work"].Command, want.Targets["work"].Command)
	}
}

func TestLoadMissingFileErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.toml")); err == nil {
		t.Fatal("Load of missing file: want error, got nil")
	}
}

func TestDefaultPathEndsWithConfigFile(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if filepath.Base(p) != "config.toml" {
		t.Errorf("DefaultPath base = %q, want config.toml", filepath.Base(p))
	}
}
