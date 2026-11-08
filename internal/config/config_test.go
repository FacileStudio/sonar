package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadKeepsPacingKeysTopLevelAndReadsFetch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sonar.yml")
	body := "minInterval: 2s\nmaxConcurrency: 3\nfetch:\n  publicOnly: true\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MinInterval != 2*time.Second || cfg.MaxConcurrency != 3 || cfg.RetryBackoff != 800*time.Millisecond {
		t.Fatalf("pacing not parsed: %+v", cfg.Pacing)
	}
	if !cfg.FetchPublicOnly() {
		t.Fatal("fetch.publicOnly ignored")
	}
}

func TestFetchPublicOnlyFromEnv(t *testing.T) {
	cfg := Default()
	t.Setenv("SONAR_FETCH_PUBLIC_ONLY", "")
	if cfg.FetchPublicOnly() {
		t.Fatal("public-only on by default")
	}
	t.Setenv("SONAR_FETCH_PUBLIC_ONLY", "1")
	if !cfg.FetchPublicOnly() {
		t.Fatal("SONAR_FETCH_PUBLIC_ONLY=1 ignored")
	}
}
