package engines

import (
	"fmt"

	"github.com/go-rod/rod"
)

func selectHTML(page *rod.Page, selector string) string {
	if selector != "" {
		if ok, el, _ := page.Has(selector); ok && el != nil {
			if h, err := el.HTML(); err == nil && h != "" {
				return h
			}
		}
	}
	return defaultContentHTML(page)
}

func defaultContentHTML(page *rod.Page) string {
	tags := []string{"article", "main", "#content", "body"}
	for _, tag := range tags {
		if ok, el, _ := page.Has(tag); ok && el != nil {
			if h, err := el.HTML(); err == nil && h != "" {
				return h
			}
		}
	}
	h, _ := page.HTML()
	return h
}

func extractContent(page *rod.Page, selector string) (string, string, error) {
	raw := selectHTML(page, selector)
	if raw == "" {
		return "", "", fmt.Errorf("%w: no content extracted", ErrTransient)
	}
	md, err := HTMLToMarkdown(raw)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", ErrTransient, err)
	}
	return pageTitle(page), md, nil
}

func pageTitle(page *rod.Page) string {
	info, err := page.Info()
	if err != nil || info == nil {
		return ""
	}
	return info.Title
}
