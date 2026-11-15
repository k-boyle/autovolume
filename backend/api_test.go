package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/k-boyle/autovolume/database"
	v1 "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/sony"
)

func TestApi(t *testing.T) {
	connection := os.Getenv("AV_TEST_DATABASE_CONNECTION")
	if connection == "" {
		t.Skip("AV_TEST_DATABASE_CONNECTION is not set")
	}

	db, err := database.New(connection)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureSettings(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	session := &database.Session{
		Id:        rand.Text(),
		Username:  "test",
		AvatarUrl: "https://example.com/avatar.png",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := db.InsertSession(t.Context(), session); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteSession(context.Background(), session.Id) })

	websocket := plex.NewWebsocket(t.Context())
	t.Cleanup(func() { _ = websocket.Close() })

	srv := httptest.NewServer(backend(
		db,
		plex.NewAuthClient(http.DefaultClient),
		plex.NewApiClient(http.DefaultClient, http.DefaultClient),
		websocket,
		sony.NewClient(http.DefaultClient),
	))
	t.Cleanup(srv.Close)

	client := connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL))
	auth := autovolumev1connect.NewAuthServiceClient(client)
	setup := autovolumev1connect.NewSetupServiceClient(client)

	_, err = auth.GetSession(t.Context(), &v1.GetSessionRequest{})
	if code := connect.CodeOf(err); code != connect.CodeUnauthenticated {
		t.Errorf("GetSession without cookie: code = %v, want %v", code, connect.CodeUnauthenticated)
	}

	ctx, info := connect.NewClientContext(t.Context())
	info.RequestHeader().Set("Cookie", "av_session="+session.Id)

	res, err := auth.GetSession(ctx, &v1.GetSessionRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if res.GetUser().GetUsername() != "test" {
		t.Errorf("username = %q, want %q", res.GetUser().GetUsername(), "test")
	}

	if _, err := setup.GetSetupStatus(ctx, &v1.GetSetupStatusRequest{}); err != nil {
		t.Fatal(err)
	}
}
