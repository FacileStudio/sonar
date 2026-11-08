package engines

import (
	"context"
	"errors"
	"testing"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
)

func TestRodName(t *testing.T) {
	r := &Rod{}
	if r.Name() != "rod" {
		t.Fatalf("expected rod, got %s", r.Name())
	}
}

func TestRodSearchInvalidURL(t *testing.T) {
	r := &Rod{}
	_, err := r.Search(context.Background(), "ftp://example.com", 1)
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("expected ErrBlocked for non-http URL, got %v", err)
	}

	_, err = r.Search(context.Background(), "invalid-query", 1)
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("expected ErrBlocked for plain text query, got %v", err)
	}
}

func TestRodWaitUnknownCondition(t *testing.T) {
	r := &Rod{WaitCondition: "invalid"}
	err := r.wait(nil)
	if err == nil {
		t.Fatalf("expected error for unknown wait condition, got nil")
	}
}

func TestRodWaitNoneCondition(t *testing.T) {
	r := &Rod{WaitCondition: "none"}
	err := r.wait(nil)
	if err != nil {
		t.Fatalf("expected nil for none wait condition, got %v", err)
	}
}

func TestNewLauncherPrefersSystemBrowser(t *testing.T) {
	bin, ok := launcher.LookPath()
	got := newLauncher("").Get(flags.Bin)
	if ok && got != bin {
		t.Fatalf("launcher bin = %q, want system browser %q", got, bin)
	}
	if !ok && got != "" {
		t.Fatalf("launcher bin = %q without a system browser", got)
	}
}
