package auth

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (a authServiceHandler) CompleteLogin(ctx context.Context, req *CompleteLoginRequest) (*CompleteLoginResponse, error) {
	login, err := a.db.GetPendingLogin(ctx, req.LoginId)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, "failed to get pending login").WithCause(err)
	}
	if login == nil {
		return nil, connect.NewError(connect.CodeNotFound, "pending login not found")
	}
	if login.ExpiresAt.Before(time.Now()) {
		go func() {
			if err := a.db.DeletePendingLogin(ctx, req.LoginId); err != nil {
				slog.Error("failed to delete pending login", slog.Any("error", err))
			}
		}()
		return nil, connect.NewError(connect.CodeUnauthenticated, "pending login has expired")
	}

	pin, err := a.plexAuth.ExchangePin(ctx, login.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to exchange auth pin").WithCause(err)
	}
	if pin.AuthToken == nil {
		return &CompleteLoginResponse{}, nil
	}

	go func() {
		if err := a.db.DeletePendingLogin(ctx, req.LoginId); err != nil {
			slog.Error("failed to delete pending login", slog.Any("error", err))
		}
	}()

	settings, err := a.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeNotFound, "settings not found")
	}

	settings.PlexToken = *pin.AuthToken
	settings.PlexTokenExpiresAt = pin.ExpiresAt

	if err := a.db.UpdatePlexAuth(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to update plex auth").WithCause(err)
	}

	users, err := a.plexApi.GetUsers.Execute(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get plex users").WithCause(err)
	}

	if len(users.Users) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, "no plex users found")
	}

	user := users.Users[0]

	session := &database.Session{
		Id:        rand.Text(),
		Username:  user.Username,
		AvatarUrl: user.AvatarUrl,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 30),
	}

	if err := a.db.InsertSession(ctx, session); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to insert new session").WithCause(err)
	}

	rsp := &CompleteLoginResponse{
		Complete: true,
		User: &User{
			Username:  user.Username,
			AvatarUrl: user.AvatarUrl,
		},
	}
	cookie := &http.Cookie{
		Name:     "av_session",
		Value:    session.Id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(30 * 24 * time.Hour),
	}
	info, _ := connect.CallInfoForServerContext(ctx)
	info.ResponseHeader().Set("Set-Cookie", cookie.String())
	return rsp, nil
}
