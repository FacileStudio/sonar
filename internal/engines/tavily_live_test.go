package engines

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/httpc"
)

func TestTavilyLive(t *testing.T) {
	key := os.Getenv("TAVILY_KEY")
	if key == "" {
		t.Skip("TAVILY_KEY not set")
	}
	tv := &Tavily{Key: key, Client: httpc.Client(20 * time.Second), Count: 3}
	res, err := tv.Search(context.Background(), "golang http client", 3)
	if err != nil {
		t.Fatalf("tavily error: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("tavily returned 0 results")
	}
	t.Logf("got %d results: %s", len(res), res[0].Title)
}
