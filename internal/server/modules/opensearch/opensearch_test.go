package opensearch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestDescriptionCarriesTheSearchAndSuggestTemplates(t *testing.T) {
	body, err := Description("https://sonar.example/")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`xmlns="http://a9.com/-/spec/opensearch/1.1/"`,
		`template="https://sonar.example/search?q={searchTerms}"`,
		`type="application/x-suggestions+json"`,
		`<ShortName>Sonar</ShortName>`,
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("description lacks %s:\n%s", want, body)
		}
	}
}

func TestSearchEntryRedirectsExternalBangsAndServesTheAppOtherwise(t *testing.T) {
	router := chi.NewRouter()
	app := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("app")) })
	RegisterRoutes(router, "https://sonar.example", app)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/search?q=%21g+go+generics", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://www.google.com/search?q=go+generics" {
		t.Fatalf("bang: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/search?q=go+generics+%21i", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "app" {
		t.Fatalf("internal bang must load the app: %d %q", rec.Code, rec.Body.String())
	}
}
