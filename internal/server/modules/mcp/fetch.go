package mcp

import (
	"context"
	"fmt"
	"unicode/utf8"

	troncerrors "github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultLength = 30_000
	maxLength     = 60_000
)

// fetch is sonar_fetch.
func (t *toolset) fetch(ctx context.Context, req *sdk.CallToolRequest, in fetchInput) (*sdk.CallToolResult,
	fetchOutput, error) {
	userID, err := userOf(req)
	if err != nil {
		return nil, fetchOutput{}, err
	}
	length, err := in.length()
	if err != nil {
		return nil, fetchOutput{}, err
	}
	page, err := t.searcher.Fetch(ctx, userID, in.URL)
	if err != nil {
		return nil, fetchOutput{}, mcpkit.Failure(err, "")
	}
	out := fetchOutput{URL: page.URL, FinalURL: page.FinalURL, Title: page.Title, TotalLength: len(page.Markdown),
		SourceTruncated: page.Truncated}
	out.window(page.Markdown, in.Offset, length)
	return nil, out, nil
}

// length checks offset and max_length and returns the window size.
func (in fetchInput) length() (int, error) {
	if in.Offset < 0 {
		return 0, troncerrors.Invalid("offset must be 0 or more, from next_offset")
	}
	switch {
	case in.MaxLength == 0:
		return defaultLength, nil
	case in.MaxLength < 1000 || in.MaxLength > maxLength:
		return 0, troncerrors.Invalid("max_length must be between 1000 and 60000")
	}
	return in.MaxLength, nil
}

// window puts the part of markdown from offset, at most length bytes, in
// out, with the offset to read on and a note on what is left.
func (out *fetchOutput) window(markdown string, offset, length int) {
	if offset >= len(markdown) && offset > 0 {
		out.Offset = len(markdown)
		out.Note = fmt.Sprintf("offset is past the end: the page has %d bytes of markdown.", len(markdown))
		return
	}
	for offset > 0 && !utf8.RuneStart(markdown[offset]) {
		offset--
	}
	part, cut := mcpkit.Clip(markdown[offset:], length)
	out.Offset, out.Markdown = offset, part
	if cut {
		out.NextOffset = offset + len(part)
		out.Note = fmt.Sprintf("Showing bytes %d to %d of %d. Call sonar_fetch again with offset=%d to read on.",
			offset, out.NextOffset, len(markdown), out.NextOffset)
	}
	if out.SourceTruncated {
		out.Note += " The page was longer than sonar reads (512 KiB); the end is missing."
	}
}
