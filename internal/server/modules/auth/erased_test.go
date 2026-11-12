package auth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
	jwtpkg "github.com/FacileStudio/porte/oidc/jwt"
	"github.com/FacileStudio/sonar/internal/server/servertest"
	"github.com/FacileStudio/tronc/mcp/mcptest"
)

func TestAnErasedSubjectIsRefusedAndARecreatedAccountIsHeldAgain(t *testing.T) {
	db := servertest.DB(t)
	ctx := context.Background()
	issuer := mcptest.NewIssuer(t)
	kits, err := New(ctx, db, Config{Porte: porte.Config{
		Issuer: issuer.URL, ClientID: "sonar", ClientSecret: "secret", SSOOnly: true, AppName: "Sonar",
		RedirectURL: "https://sonar.example/api/auth/oidc/callback", SuccessURL: "https://sonar.example/",
	}, ClientID: "sonar", ClientSecret: "secret"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	gone := personWithData(t, db, "gone@example.com", issuer.URL, "sub-gone")
	kept := personWithData(t, db, "kept@example.com", issuer.URL, "sub-kept")
	deletions := kits.Deletions
	if err := deletions.OnPersonDeleted(ctx, deletions.Client.Issuer, "sub-gone"); err != nil {
		t.Fatal(err)
	}
	if err := deletions.Erase(ctx, deletions.Client.Issuer, "sub-gone"); err != nil {
		t.Fatal(err)
	}
	state, _ := json.Marshal(map[string]any{"cursor": time.Now(), "done": map[string]time.Time{"sub-gone": time.Now()}})
	if err := deletions.State.Save(ctx, state); err != nil {
		t.Fatal(err)
	}

	verified := true
	claims := jwtpkg.Claims{Subject: "sub-gone", Email: "gone@example.com"}
	claims.Profile.EmailVerified = &verified
	if _, err := kits.OIDC.ProvisionBearer(ctx, claims); !errors.Is(err, porte.ErrSubjectErased) {
		t.Fatalf("provisioning an erased subject: %v", err)
	}
	if n := rowsOf(t, db, gone); n != 0 {
		t.Fatalf("the refused provisioning left %d rows", n)
	}

	if held, err := deletions.Holds(ctx, deletions.Client.Issuer, "sub-gone"); err != nil || held {
		t.Fatalf("an erased subject is held: %v %v", held, err)
	}
	personWithData(t, db, "gone@example.com", issuer.URL, "sub-gone")
	if held, err := deletions.Holds(ctx, deletions.Client.Issuer, "sub-gone"); err != nil || !held {
		t.Fatalf("a re-created account is not held: %v %v", held, err)
	}
	if n := rowsOf(t, db, kept); n != 1+len(personTables) {
		t.Fatalf("the other person kept %d rows, want %d", n, 1+len(personTables))
	}
}
