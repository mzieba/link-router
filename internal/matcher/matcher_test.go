package matcher

import (
	"testing"

	"github.com/mzieba/link-router/internal/config"
)

func TestMatchHostGlob(t *testing.T) {
	rules := []config.Rule{{MatchHost: []string{"*.example.com"}, Target: "work"}}
	got, ok, err := Match(rules, "https://app.example.com/path")
	if err != nil || !ok || got != "work" {
		t.Fatalf("Match = (%q, %v, %v), want (work, true, nil)", got, ok, err)
	}
}

func TestMatchURLRegex(t *testing.T) {
	rules := []config.Rule{{MatchURL: `^https://github\.com/acme/`, Target: "work"}}
	got, ok, _ := Match(rules, "https://github.com/acme/repo")
	if !ok || got != "work" {
		t.Fatalf("Match = (%q, %v), want (work, true)", got, ok)
	}
}

func TestMatchRequiresAllPresentMatchers(t *testing.T) {
	rules := []config.Rule{{
		MatchHost: []string{"github.com"},
		MatchURL:  `/acme/`,
		Target:    "work",
	}}
	if _, ok, _ := Match(rules, "https://github.com/other/repo"); ok {
		t.Fatal("host matched but url did not; want no match")
	}
	if _, ok, _ := Match(rules, "https://github.com/acme/repo"); !ok {
		t.Fatal("both matchers satisfied; want match")
	}
}

func TestMatchFirstWins(t *testing.T) {
	rules := []config.Rule{
		{MatchHost: []string{"*.example.com"}, Target: "first"},
		{MatchHost: []string{"*.example.com"}, Target: "second"},
	}
	got, _, _ := Match(rules, "https://a.example.com")
	if got != "first" {
		t.Fatalf("Match = %q, want first", got)
	}
}

func TestMatchNoMatchAndEmptyRuleSkipped(t *testing.T) {
	rules := []config.Rule{{Target: "empty"}} // no matchers -> never matches
	if _, ok, _ := Match(rules, "https://a.example.com"); ok {
		t.Fatal("rule with no matchers matched; want no match")
	}
}

func TestMatchInvalidRegexErrors(t *testing.T) {
	rules := []config.Rule{{MatchURL: "(", Target: "x"}}
	if _, _, err := Match(rules, "https://a.example.com"); err == nil {
		t.Fatal("invalid regex: want error")
	}
}

func TestMatchInvalidHostGlobErrors(t *testing.T) {
	rules := []config.Rule{{MatchHost: []string{"["}, Target: "x"}}
	if _, _, err := Match(rules, "https://a.example.com"); err == nil {
		t.Fatal("invalid host glob: want error")
	}
}
