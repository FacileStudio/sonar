package quota

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/server/servertest"
)

func TestTakeStopsAtTheLimitUnderConcurrency(t *testing.T) {
	db := servertest.DB(t)
	user := servertest.User(t, db, "a@example.com", "https://sso.example", "sub-a")
	counter := &Counter{DB: db}
	var wg sync.WaitGroup
	var mu sync.Mutex
	granted := 0
	for range 40 {
		wg.Go(func() {
			ok, err := counter.Take(context.Background(), user, 20)
			if err != nil {
				t.Error(err)
			}
			if ok {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if granted != 20 {
		t.Fatalf("granted %d searches, want 20", granted)
	}
	if used, _ := counter.Used(context.Background(), user); used != 20 {
		t.Fatalf("used = %d", used)
	}
}

func TestZeroLimitRefusesAndUnlimitedCounts(t *testing.T) {
	db := servertest.DB(t)
	user := servertest.User(t, db, "b@example.com", "https://sso.example", "sub-b")
	counter := &Counter{DB: db}
	if ok, _ := counter.Take(context.Background(), user, 0); ok {
		t.Fatal("a zero limit must refuse")
	}
	for range 3 {
		if ok, _ := counter.Take(context.Background(), user, -1); !ok {
			t.Fatal("unlimited must grant")
		}
	}
	if err := counter.Refund(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	if used, _ := counter.Used(context.Background(), user); used != 2 {
		t.Fatalf("used = %d, want 2", used)
	}
}

func TestTheCounterResetsAtUTCMidnight(t *testing.T) {
	db := servertest.DB(t)
	user := servertest.User(t, db, "c@example.com", "https://sso.example", "sub-c")
	now := time.Date(2026, 11, 8, 23, 59, 0, 0, time.UTC)
	counter := &Counter{DB: db, Now: func() time.Time { return now }}
	if ok, _ := counter.Take(context.Background(), user, 1); !ok {
		t.Fatal("first search refused")
	}
	if ok, _ := counter.Take(context.Background(), user, 1); ok {
		t.Fatal("second search granted over the limit")
	}
	now = now.Add(2 * time.Minute)
	if ok, _ := counter.Take(context.Background(), user, 1); !ok {
		t.Fatal("the next day must start a fresh counter")
	}
	now = now.AddDate(0, 0, 3)
	if swept, err := counter.Sweep(context.Background()); err != nil || swept != 2 {
		t.Fatalf("swept %d, %v", swept, err)
	}
}
