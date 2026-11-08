package engines

import (
	"os"
	"strings"
	"testing"

	"github.com/go-rod/rod/lib/launcher"
)

func requireBrowser(t *testing.T) {
	t.Helper()
	if os.Getenv("SONAR_SKIP_BROWSER_TESTS") != "" {
		t.Skip("SONAR_SKIP_BROWSER_TESTS is set")
	}
	if _, ok := launcher.LookPath(); ok {
		return
	}
	if err := launcher.NewBrowser().Validate(); err != nil {
		t.Skip("no local chromium")
	}
}

func TestGuardedBrowserRefusesInternalNavigation(t *testing.T) {
	requireBrowser(t)
	base := trapServer(t)
	for _, path := range []string{"/302", "/js", "/meta", "/frame", "/ws"} {
		requireBlocked(t, extractGuarded(t, base+path, true))
	}
}

func TestGuardedBrowserAllowsPublicPage(t *testing.T) {
	requireBrowser(t)
	o := extractGuarded(t, trapServer(t)+"/ok", true)
	if o.Err != nil || !strings.Contains(o.Result.Snippet, "public page") || o.Result.Engine != "rod" {
		t.Fatalf("public page via browser: %v %q", o.Err, o.Result.Snippet)
	}
}
