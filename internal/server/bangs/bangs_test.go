package bangs

import (
	"strings"
	"testing"
)

func TestParseFindsBangsAnywhere(t *testing.T) {
	cases := []struct {
		query, rest, trigger string
		ok                   bool
	}{
		{"!g golang generics", "golang generics", "g", true},
		{"golang generics !g", "golang generics", "g", true},
		{"cats !I", "cats", "i", true},
		{"price of !unknown thing", "price of !unknown thing", "", false},
		{"no bang here", "no bang here", "", false},
		{"!", "!", "", false},
	}
	for _, c := range cases {
		rest, bang, ok := Parse(c.query)
		if rest != c.rest || ok != c.ok || bang.Trigger != c.trigger {
			t.Errorf("Parse(%q) = %q, %q, %v", c.query, rest, bang.Trigger, ok)
		}
	}
}

func TestExternalTargetEscapesTheQuery(t *testing.T) {
	_, bang, _ := Parse("!g a&b c")
	got := bang.Target("a&b c")
	if got != "https://www.google.com/search?q=a%26b+c" {
		t.Fatalf("target = %q", got)
	}
}

func TestEveryBangIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range List() {
		set := 0
		for _, v := range []string{b.Kind, b.Engine, b.URL} {
			if v != "" {
				set++
			}
		}
		if set != 1 || seen[b.Trigger] || b.Trigger != strings.ToLower(b.Trigger) {
			t.Errorf("bad bang %+v", b)
		}
		if b.External() && (!strings.HasPrefix(b.URL, "https://") || strings.Count(b.URL, "%s") != 1) {
			t.Errorf("external bang %q must be https with one %%s", b.Trigger)
		}
		seen[b.Trigger] = true
	}
}
