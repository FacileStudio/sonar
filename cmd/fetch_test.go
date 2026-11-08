package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/netguard"
)

func TestFetchCommandValidation(t *testing.T) {
	err := runFetch(fetchCmd, []string{"ftp://invalid.com"})
	if err == nil || !strings.Contains(err.Error(), "URL must start with http") {
		t.Fatalf("expected URL prefix error, got %v", err)
	}

	err = runFetch(fetchCmd, []string{"example.com"})
	if err == nil || !strings.Contains(err.Error(), "URL must start with http") {
		t.Fatalf("expected URL prefix error, got %v", err)
	}
}

func TestFetchCommandRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "fetch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected fetch command to be registered on rootCmd")
	}
}

func TestFetchJSONReportsBlockedCode(t *testing.T) {
	err := errors.New("boom")
	got := toFetchJSON(engines.Outcome{URL: "http://x.test/", Err: &netguard.BlockedError{Code: netguard.CodeInternal, Reason: "r"}})
	if got.Error == nil || got.Error.Code != netguard.CodeInternal || got.Markdown != "" {
		t.Fatalf("blocked outcome: %+v", got)
	}
	if got := toFetchJSON(engines.Outcome{URL: "u", Err: err}); got.Error.Code != "fetch_failed" {
		t.Fatalf("plain failure: %+v", got)
	}
}

func TestFetchPublicOnlyRefusesMetadataEndpoint(t *testing.T) {
	flagPublicOnly, flagFetchJSON = true, true
	defer func() { flagPublicOnly, flagFetchJSON = false, false }()
	err := runFetch(fetchCmd, []string{"http://169.254.169.254/latest/meta-data/"})
	if b, ok := netguard.AsBlocked(err); !ok || b.Code != netguard.CodeInternal {
		t.Fatalf("metadata endpoint not refused: %v", err)
	}
}
