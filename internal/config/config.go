package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the full parsed configuration.
type Config struct {
	Default string            `toml:"default"`
	Rules   []Rule            `toml:"rules"`
	Targets map[string]Target `toml:"targets"`
}

// Rule maps matching URLs to a target. It matches when it has at least one
// matcher and all present matchers match.
type Rule struct {
	MatchHost []string `toml:"match_host,omitempty"`
	MatchURL  string   `toml:"match_url,omitempty"`
	Target    string   `toml:"target"`
}

// Target is a browser command with its arguments. The token {url} in Args is
// replaced with the incoming URL at launch time.
type Target struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

// Load reads and parses the config file at path.
func Load(path string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	return &c, nil
}

// Save writes c to path as TOML, creating parent directories as needed.
func Save(path string, c *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config %s: %w", path, err)
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		_ = f.Close()
		return fmt.Errorf("encode config: %w", err)
	}
	// Report close errors: they surface write failures that Encode cannot see.
	if err := f.Close(); err != nil {
		return fmt.Errorf("close config %s: %w", path, err)
	}
	return nil
}

// DefaultPath returns the OS-appropriate config file path.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(dir, "link-router", "config.toml"), nil
}

// Sample returns a starter configuration. Its values mirror the example config
// in the README so a fresh `init` produces what the documentation shows.
func Sample() *Config {
	return &Config{
		Default: "personal",
		Rules: []Rule{
			{MatchHost: []string{"*.my-company.com", "my-company.com", "jira.*"}, Target: "work"},
			{MatchURL: `^https://github\.com/my-company`, Target: "work"},
		},
		Targets: map[string]Target{
			"work":     {Command: "google-chrome", Args: []string{"--profile-directory=Work", "{url}"}},
			"personal": {Command: "firefox", Args: []string{"-P", "personal", "{url}"}},
		},
	}
}
