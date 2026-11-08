package engines

import (
	"fmt"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type docStatus struct {
	mu     sync.Mutex
	status int
}

func watchDocument(page *rod.Page) *docStatus {
	ds := &docStatus{}
	frame := page.FrameID
	go page.EachEvent(func(e *proto.NetworkResponseReceived) {
		if e.Type == proto.NetworkResourceTypeDocument && e.FrameID == frame && e.Response != nil {
			ds.mu.Lock()
			ds.status = e.Response.Status
			ds.mu.Unlock()
		}
	})()
	return ds
}

func (ds *docStatus) get() int {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.status
}

func pageURL(page *rod.Page, fallback string) string {
	info, err := page.Info()
	if err != nil || info == nil || info.URL == "" {
		return fallback
	}
	return info.URL
}

func (r *Rod) render(page *rod.Page, query string, guard *rodGuard) ([]Result, error) {
	doc := watchDocument(page)
	if err := r.load(page, query, guard); err != nil {
		return nil, err
	}
	title, snippet, err := extractContent(page, r.Selector)
	if verdict := guard.verdict(); verdict != nil {
		return nil, verdict
	}
	if err != nil {
		return nil, err
	}
	if title == "" {
		title = query
	}
	return []Result{{
		Title: title, Snippet: snippet, URL: query, Engine: "rod", Priority: 12,
		FinalURL: pageURL(page, query), Status: doc.get(),
	}}, nil
}

func (r *Rod) load(page *rod.Page, query string, guard *rodGuard) error {
	navErr := page.Navigate(query)
	if navErr == nil {
		_ = r.wait(page)
	}
	if err := guard.verdict(); err != nil {
		return err
	}
	if navErr != nil {
		return fmt.Errorf("%w: %s", ErrTransient, navErr)
	}
	return nil
}
