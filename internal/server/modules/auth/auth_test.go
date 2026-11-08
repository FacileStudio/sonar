package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/sonar/internal/server/servertest"
)

func TestPictureUsersSetsAvatarFromPicture(t *testing.T) {
	db := servertest.DB(t)
	users := pictureUsers{UserStore: portepg.New(db).Users()}
	ctx := context.Background()
	claims := porte.Claims{
		Provider: "https://sso.example",
		Subject:  "sub-1",
		Email:    "pic@example.com",
		Picture:  "https://sso.example/avatars/sub-1",
	}
	id, err := users.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatal(err)
	}
	var avatarURL string
	if err := db.QueryRow(`SELECT avatar_url FROM porte_users WHERE id = $1`, id).Scan(&avatarURL); err != nil {
		t.Fatal(err)
	}
	if avatarURL != "https://sso.example/avatars/sub-1" {
		t.Fatalf("avatar_url = %q, want %q", avatarURL, "https://sso.example/avatars/sub-1")
	}
}

func TestPictureUsersPreservesExistingAvatar(t *testing.T) {
	db := servertest.DB(t)
	users := pictureUsers{UserStore: portepg.New(db).Users()}
	ctx := context.Background()
	claims := porte.Claims{
		Provider:  "https://sso.example",
		Subject:   "sub-2",
		Email:     "custom@example.com",
		AvatarURL: "https://custom.example/avatar.png",
		Picture:   "https://sso.example/avatars/sub-2",
	}
	id, err := users.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatal(err)
	}
	var avatarURL string
	if err := db.QueryRow(`SELECT avatar_url FROM porte_users WHERE id = $1`, id).Scan(&avatarURL); err != nil {
		t.Fatal(err)
	}
	if avatarURL != "https://custom.example/avatar.png" {
		t.Fatalf("avatar_url = %q, want %q", avatarURL, "https://custom.example/avatar.png")
	}
}

func TestBackfillPictures(t *testing.T) {
	db := servertest.DB(t)
	ctx := context.Background()
	issuer := "https://sso.example"
	user := servertest.User(t, db, "backfill@example.com", issuer, "sub-backfill")
	filled, err := backfillPictures(ctx, db, issuer)
	if err != nil {
		t.Fatal(err)
	}
	if filled != 1 {
		t.Fatalf("filled = %d, want 1", filled)
	}
	var avatarURL, avatarSource string
	if err := db.QueryRow(`SELECT avatar_url, avatar_source FROM porte_users WHERE id = $1`, user).
		Scan(&avatarURL, &avatarSource); err != nil {
		t.Fatal(err)
	}
	want := issuer + "/avatars/sub-backfill"
	if avatarURL != want || avatarSource != "oidc" {
		t.Fatalf("got avatar_url=%q, source=%q; want %q, oidc", avatarURL, avatarSource, want)
	}
	again, err := backfillPictures(ctx, db, issuer)
	if err != nil || again != 0 {
		t.Fatalf("second run returned (%d, %v), want (0, nil)", again, err)
	}
}

func TestHandleSessionFallback(t *testing.T) {
	db := servertest.DB(t)
	issuer := "https://sso.example"
	user := servertest.User(t, db, "session@example.com", issuer, "sub-session")
	handler := handleSession(db, Features{}, issuer)
	req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)
	req = req.WithContext(porte.WithIdentity(req.Context(), porte.Identity{
		UserID: user,
		Email:  "session@example.com",
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var res sessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.User == nil || res.User.AvatarURL != issuer+"/avatars/sub-session" {
		t.Fatalf("unexpected user response: %+v", res.User)
	}
}
