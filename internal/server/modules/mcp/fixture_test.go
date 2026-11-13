package mcp

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/mcpauth"
	"github.com/FacileStudio/sonar/internal/server/imgsig"
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	"github.com/FacileStudio/sonar/websearch"
	troncerrors "github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/go-chi/chi/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const subject = "person-sub"

type identities struct{ porte.IdentityStore }

func (identities) Find(_ context.Context, provider, sub string) (porte.StoredIdentity, error) {
	if sub != subject {
		return porte.StoredIdentity{}, porte.ErrNotFound
	}
	return porte.StoredIdentity{UserID: 7, Provider: provider, Subject: sub}, nil
}

type engines struct {
	name     string
	queries  []string
	markdown string
	fail     error
}

func (e *engines) Search(_ context.Context, q string, _ int, _ []string) ([]websearch.Result, error) {
	e.queries = append(e.queries, q)
	return []websearch.Result{{Title: "hit", URL: "https://r.example/", Snippet: "s", Engine: e.name}}, nil
}

func (e *engines) Images(context.Context, string, int) ([]websearch.Image, error) {
	return []websearch.Image{{Title: "img", URL: "https://p.example/", Thumbnail: "https://imgs.example/t.jpg"}}, nil
}

func (e *engines) News(context.Context, string, int) ([]websearch.News, error) { return nil, nil }

func (e *engines) Fetch(_ context.Context, rawURL string, _ websearch.FetchOptions) (*websearch.Page, error) {
	if e.fail != nil {
		return nil, e.fail
	}
	return &websearch.Page{URL: rawURL, FinalURL: rawURL, Title: "Page", Markdown: e.markdown}, nil
}

type plans struct{ limit int64 }

func (p plans) Daily(context.Context, int64) (int64, error) { return p.limit, nil }
func (p plans) CheckAI(context.Context, int64) error        { return nil }
func (p plans) Refusal(message string, cause error) error {
	return troncerrors.New("payment_required", message, cause).WithUpgradeURL("https://sso.example/billing")
}

type quota struct{ used int64 }

func (q *quota) Take(_ context.Context, _, limit int64) (bool, error) {
	if limit >= 0 && q.used >= limit {
		return false, nil
	}
	q.used++
	return true, nil
}
func (q *quota) Refund(context.Context, int64) error        { q.used--; return nil }
func (q *quota) Used(context.Context, int64) (int64, error) { return q.used, nil }
func (q *quota) ResetsAt() time.Time                        { return time.Time{} }

type keys map[string]string

func (k keys) Keys(context.Context, int64) (map[string]string, error) { return k, nil }

type fixture struct {
	resource string
	issuer   *mcptest.Issuer
	instance *engines
	own      *engines
	quota    *quota
}

func newFixture(t *testing.T, limit int64, own keys) *fixture {
	t.Helper()
	signer, _ := imgsig.New(nil)
	f := &fixture{issuer: mcptest.NewIssuer(t), instance: &engines{name: "brave"}, own: &engines{name: "exa"},
		quota: &quota{}}
	service := &search.Service{Instance: f.instance, Plans: plans{limit: limit}, Quota: f.quota, Keys: own,
		Signer: signer, Logger: slog.Default(),
		ForKeys: func(map[string]string) (search.Engines, error) { return f.own, nil }}
	router := chi.NewRouter()
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	f.resource = srv.URL + "/mcp"
	verifier, err := mcpauth.Verifier(t.Context(), mcpauth.Config{Issuer: f.issuer.URL, Resource: f.resource,
		ScopePrefix: "sonar", Identities: identities{}})
	if err != nil {
		t.Fatal(err)
	}
	mcpService, err := NewService(service, mcpkit.AppConfig{Resource: f.resource, Issuer: f.issuer.URL, Version: "test",
		Verifier: verifier, PerMinute: 100_000})
	if err != nil {
		t.Fatal(err)
	}
	RegisterRoutes(router, mcpService)
	return f
}

func (f *fixture) token(t testing.TB, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = []string{"sonar.read"}
	}
	return f.issuer.Token(t, f.resource, subject, scopes...)
}

func (f *fixture) call(t *testing.T, tool string, args map[string]any) *sdk.CallToolResult {
	t.Helper()
	session := mcptest.Connect(t, f.resource, f.token(t), mcptest.ClientOptions{})
	return mcptest.Call(t, session, tool, args)
}

func wantError(t *testing.T, result *sdk.CallToolResult, part string) {
	t.Helper()
	if text := mcptest.Text(result); !result.IsError || !strings.Contains(text, part) {
		t.Fatalf("want an error result naming %q, got %v %q", part, result.IsError, text)
	}
}
