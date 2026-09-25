package engines

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/httpc"
	"github.com/FacileStudio/sonar/internal/tiroir"
)

func parallelTestKey() string {
	if k := os.Getenv("PARALLEL_API_KEY"); k != "" {
		return k
	}
	if k := os.Getenv("SONAR_PARALLEL_KEY"); k != "" {
		return k
	}
	return tiroir.Get("PARALLEL_API_KEY")
}

func TestParallelLive(t *testing.T) {
	key := parallelTestKey()
	if key == "" {
		t.Skip("parallel key not set")
	}
	p := &Parallel{Key: key, Client: httpc.Client(20 * time.Second), Count: 3}
	res, err := p.Search(context.Background(), "golang http client", 3)
	if err != nil {
		t.Fatalf("parallel search error: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("parallel returned 0 results")
	}
	if res[0].Engine != "parallel" {
		t.Fatalf("expected engine parallel, got %s", res[0].Engine)
	}
}

func TestParallelExtractLive(t *testing.T) {
	key := parallelTestKey()
	if key == "" {
		t.Skip("parallel key not set")
	}
	p := &Parallel{Key: key, Client: httpc.Client(20 * time.Second), Count: 1}
	res, err := p.Extract(context.Background(), []string{"https://go.dev"})
	if err != nil {
		t.Fatalf("parallel extract error: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("parallel extract returned 0 results")
	}
	if len(res[0].Snippet) < 100 {
		t.Fatalf("expected meaningful content, got %d bytes", len(res[0].Snippet))
	}
}
