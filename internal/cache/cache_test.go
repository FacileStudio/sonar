package cache

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/engines"
)

func TestLongQueriesWithSharedPrefixDoNotCollide(t *testing.T) {
	c, err := New(t.TempDir(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	prefix := strings.Repeat("alpha ", 30)
	a := prefix + "first"
	b := prefix + "second"
	if c.path("brave", a) == c.path("brave", b) {
		t.Fatal("distinct long queries with a shared prefix produced the same cache path")
	}
	if err := c.Put("brave", a, []engines.Result{{Title: "first"}}); err != nil {
		t.Fatal(err)
	}
	if err := c.Put("brave", b, []engines.Result{{Title: "second"}}); err != nil {
		t.Fatal(err)
	}
	gotA, ok := c.Get("brave", a, 10)
	if !ok || len(gotA) != 1 || gotA[0].Title != "first" {
		t.Fatalf("query A read back %v (ok=%v), want its own result", gotA, ok)
	}
	gotB, ok := c.Get("brave", b, 10)
	if !ok || len(gotB) != 1 || gotB[0].Title != "second" {
		t.Fatalf("query B read back %v (ok=%v), want its own result", gotB, ok)
	}
}

func TestConcurrentGetPut(t *testing.T) {
	c, err := New(t.TempDir(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(2)
		go readRepeatedly(t, c, fmt.Sprintf("query %d", i), &wg)
		go putRepeatedly(t, c, fmt.Sprintf("query %d", i), &wg)
	}
	wg.Wait()
}

func readRepeatedly(t *testing.T, c *Cache, query string, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	for range 50 {
		c.Get("brave", query, 10)
	}
}

func putRepeatedly(t *testing.T, c *Cache, query string, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	for range 50 {
		results := []engines.Result{{Title: query}}
		if err := c.Put("brave", query, results); err != nil {
			t.Error(err)
		}
	}
}
