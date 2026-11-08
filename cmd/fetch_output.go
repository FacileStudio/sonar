package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/netguard"
)

type fetchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type fetchJSON struct {
	URL      string      `json:"url"`
	Title    string      `json:"title,omitempty"`
	Markdown string      `json:"markdown,omitempty"`
	Engine   string      `json:"engine,omitempty"`
	Error    *fetchError `json:"error,omitempty"`
}

func describeFetchError(err error) *fetchError {
	if b, ok := netguard.AsBlocked(err); ok {
		return &fetchError{Code: b.Code, Message: b.Reason}
	}
	return &fetchError{Code: "fetch_failed", Message: err.Error()}
}

func toFetchJSON(o engines.Outcome) fetchJSON {
	if o.Err != nil {
		return fetchJSON{URL: o.URL, Error: describeFetchError(o.Err)}
	}
	return fetchJSON{URL: o.URL, Title: o.Result.Title, Markdown: o.Result.Snippet, Engine: o.Result.Engine}
}

func allFailed(outcomes []engines.Outcome) error {
	for _, o := range outcomes {
		if o.Err == nil {
			return nil
		}
	}
	return outcomes[0].Err
}

func printFetchJSON(outcomes []engines.Outcome) error {
	rows := make([]fetchJSON, len(outcomes))
	for i, o := range outcomes {
		rows[i] = toFetchJSON(o)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rows); err != nil {
		return err
	}
	return allFailed(outcomes)
}

func printFetchMarkdown(outcomes []engines.Outcome) error {
	for _, o := range outcomes {
		switch {
		case o.Err != nil:
			fmt.Fprintf(os.Stderr, "sonar: %s: %v\n", o.URL, o.Err)
		case len(outcomes) > 1:
			fmt.Printf("# %s\n\n%s\n\n", o.URL, o.Result.Snippet)
		default:
			fmt.Println(o.Result.Snippet)
		}
	}
	return allFailed(outcomes)
}
