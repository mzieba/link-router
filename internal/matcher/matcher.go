package matcher

import (
	"fmt"
	"net/url"
	"path"
	"regexp"

	"github.com/mzieba/link-router/internal/config"
)

// Match evaluates rules top-to-bottom and returns the target of the first rule
// that matches rawURL. It returns ("", false, nil) when no rule matches. A rule
// matches only when it has at least one matcher and every present matcher matches.
func Match(rules []config.Rule, rawURL string) (string, bool, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false, fmt.Errorf("parse url %q: %w", rawURL, err)
	}
	host := u.Hostname()

	for _, r := range rules {
		if len(r.MatchHost) == 0 && r.MatchURL == "" {
			continue
		}
		if len(r.MatchHost) > 0 {
			ok, err := matchHost(r.MatchHost, host)
			if err != nil {
				return "", false, err
			}
			if !ok {
				continue
			}
		}
		if r.MatchURL != "" {
			ok, err := regexp.MatchString(r.MatchURL, rawURL)
			if err != nil {
				return "", false, fmt.Errorf("rule regexp %q: %w", r.MatchURL, err)
			}
			if !ok {
				continue
			}
		}
		return r.Target, true, nil
	}
	return "", false, nil
}

func matchHost(globs []string, host string) (bool, error) {
	for _, g := range globs {
		ok, err := path.Match(g, host)
		if err != nil {
			return false, fmt.Errorf("rule host glob %q: %w", g, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
