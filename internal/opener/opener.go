package opener

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mzieba/link-router/internal/config"
	"github.com/mzieba/link-router/internal/logging"
	"github.com/mzieba/link-router/internal/matcher"
)

// Runner launches a command. It must not wait for the process to exit.
type Runner func(command string, args []string) error

// DefaultRunner starts the command detached and returns immediately.
func DefaultRunner(command string, args []string) error {
	return exec.Command(command, args...).Start()
}

// Open resolves the target for rawURL and launches it via run. If the resolved
// target fails and it is not already the default, it falls back to the default
// target. All failures are logged.
func Open(c *config.Config, rawURL string, run Runner, log logging.Logger) error {
	name, matched, err := matcher.Match(c.Rules, rawURL)
	if err != nil {
		log.Errorf("match %q: %v", rawURL, err)
	}
	if !matched {
		name = c.Default
	}

	if err := launch(c, name, rawURL, run); err != nil {
		log.Errorf("launch target %q for %q: %v", name, rawURL, err)
		if name != c.Default {
			log.Infof("falling back to default target %q", c.Default)
			if err2 := launch(c, c.Default, rawURL, run); err2 != nil {
				log.Errorf("launch default target %q: %v", c.Default, err2)
				return fmt.Errorf("open %q: %w", rawURL, err2)
			}
			return nil
		}
		return fmt.Errorf("open %q: %w", rawURL, err)
	}
	return nil
}

// Resolution describes what Open would do for a URL without launching anything.
type Resolution struct {
	Target  string   // resolved target name
	Matched bool     // true if a rule matched; false means the default was used
	Command string   // the target's browser command
	Args    []string // launch args with the {url} token substituted
}

// Resolve reports the target and command Open would use for rawURL without
// launching it. The runtime fallback to the default target (which only happens
// when a launch fails) is not reflected here.
func Resolve(c *config.Config, rawURL string) (Resolution, error) {
	name, matched, err := matcher.Match(c.Rules, rawURL)
	if err != nil {
		return Resolution{}, err
	}
	if !matched {
		name = c.Default
	}
	res := Resolution{Target: name, Matched: matched}
	command, args, err := resolveTarget(c, name, rawURL)
	if err != nil {
		return res, err
	}
	res.Command = command
	res.Args = args
	return res, nil
}

// resolveTarget looks up name in c.Targets and returns its command with the
// launch args, {url} token substituted.
func resolveTarget(c *config.Config, name, rawURL string) (string, []string, error) {
	t, ok := c.Targets[name]
	if !ok {
		return "", nil, fmt.Errorf("unknown target %q", name)
	}
	if t.Command == "" {
		return "", nil, fmt.Errorf("target %q has no command", name)
	}
	return t.Command, substitute(t.Args, rawURL), nil
}

func launch(c *config.Config, name, rawURL string, run Runner) error {
	command, args, err := resolveTarget(c, name, rawURL)
	if err != nil {
		return err
	}
	return run(command, args)
}

// substitute replaces the {url} token in each arg. If no arg contains {url},
// the URL is appended as a final argument.
func substitute(args []string, rawURL string) []string {
	out := make([]string, 0, len(args)+1)
	replaced := false
	for _, a := range args {
		if strings.Contains(a, "{url}") {
			out = append(out, strings.ReplaceAll(a, "{url}", rawURL))
			replaced = true
		} else {
			out = append(out, a)
		}
	}
	if !replaced {
		out = append(out, rawURL)
	}
	return out
}
