package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/tronc/httpx"
)

func TestRequestLogsNeverCarryTheQuery(t *testing.T) {
	redactQueries()
	var logs bytes.Buffer
	router := httpx.NewRouter(httpx.Config{Logger: slog.New(slog.NewJSONHandler(&logs, nil))})
	router.Get("/*", func(w http.ResponseWriter, _ *http.Request) {})
	for _, target := range []string{
		"/api/search?q=my+secret+diagnosis&type=web",
		"/search?q=my+secret+diagnosis",
		"/api/suggest?q=my+secret",
		"/img?u=https%3A%2F%2Fimgs.example%2Fmy-secret.jpg&s=sig",
	} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, target, nil))
	}
	if strings.Contains(logs.String(), "secret") {
		t.Fatalf("a query reached the request log:\n%s", logs.String())
	}
	if !strings.Contains(logs.String(), "redacted") {
		t.Fatalf("expected redacted query fields:\n%s", logs.String())
	}
}
