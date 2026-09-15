package cmd

import (
	"strings"
	"testing"
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
