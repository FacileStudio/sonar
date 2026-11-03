package websearch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func poisonEnv(t *testing.T) (marker, home string) {
	t.Helper()
	bin := t.TempDir()
	home = t.TempDir()
	marker = filepath.Join(bin, "tiroir-called")
	script := "#!/bin/sh\ntouch " + marker + "\necho leaked\n"
	if err := os.WriteFile(filepath.Join(bin, "tiroir"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", home)
	t.Setenv("SONAR_CONFIG", filepath.Join(home, "sonar.yml"))
	for _, name := range []string{"SONAR_TAVILY_KEY", "TAVILY_KEY", "SONAR_BRAVE_KEY", "PARALLEL_API_KEY", "EXA_KEY"} {
		t.Setenv(name, "from-env")
	}
	return marker, home
}

func TestEmptyConfigReadsNoEnvAndNoTiroir(t *testing.T) {
	marker, home := poisonEnv(t)
	c, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Engines(); len(got) != 0 {
		t.Fatalf("engines enabled from outside Config: %v", got)
	}
	if _, err := c.Search(context.Background(), "q", 3, nil); !errors.Is(err, ErrNoEngines) {
		t.Fatalf("Search error = %v, want ErrNoEngines", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("tiroir was invoked")
	}
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Fatalf("home touched: %v", entries)
	}
}

func TestKeysEnableOnlyNamedEngines(t *testing.T) {
	c, err := New(Config{Keys: map[string]string{"tavily": "k", "exa": ""}})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Engines(); len(got) != 1 || got[0] != "tavily" {
		t.Fatalf("engines = %v, want [tavily]", got)
	}
	if _, err := New(Config{Keys: map[string]string{"bing": "k"}}); err == nil {
		t.Fatal("a key for a keyless engine was accepted")
	}
}
