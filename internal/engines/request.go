// Package files: shared HTTP plumbing for the keyed search APIs, which all
// speak JSON to one provider endpoint and classify failures the same way.
package engines

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// guardStatus maps a keyed API's HTTP status onto the engine error sentinels:
// authorization failures (and any extraBlocked code, e.g. linkup's 402) are a
// block, rate-limits and server errors transient, and any other non-200 an
// unwrapped error. 408 is folded into transient because every engine already
// reached ErrTransient for it via the catch-all.
func guardStatus(c int, extraBlocked int) error {
	switch {
	case c == http.StatusUnauthorized || c == http.StatusForbidden || c == extraBlocked:
		return ErrBlocked
	case c == http.StatusTooManyRequests || c == http.StatusRequestTimeout || c >= 500:
		return ErrTransient
	case c != http.StatusOK:
		return fmt.Errorf("%w: status %d", ErrTransient, c)
	}
	return nil
}

// decode runs a built request through the shared client, closes the response
// body and JSON-decodes it as out. A client failure, a guarded status or a
// decode error all come back as an error so the caller never touches the body.
func decode[T any](client *http.Client, req *http.Request, out *T, extraBlocked int) error {
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	if gerr := guardStatus(res.StatusCode, extraBlocked); gerr != nil {
		return gerr
	}
	if derr := json.NewDecoder(res.Body).Decode(out); derr != nil {
		return fmt.Errorf("%w: decode: %v", ErrTransient, derr)
	}
	return nil
}
