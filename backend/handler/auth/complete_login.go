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
	"github.com/k-boyle/autovolume/plex"
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
			if err := a.db.DeletePendingLogin(context.Background(), req.LoginId); err != nil {
				slog.Error("failed to delete pending login", slog.Any("error", err))
			}
		}()
		return nil, connect.NewError(connect.CodeUnauthenticated, "pending login has expired")
	}

	credentials, err := plex.CredentialsFromSeed(login.ClientId, login.KeyId, login.KeySeed)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read plex credentials").WithCause(err)
	}

	pin, err := a.plexAuth.ExchangePin(ctx, credentials, login.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to exchange auth pin").WithCause(err)
	}
	if pin.AuthToken == nil {
		return &CompleteLoginResponse{}, nil
	}

	go func() {
		if err := a.db.DeletePendingLogin(context.Background(), req.LoginId); err != nil {
			slog.Error("failed to delete pending login", slog.Any("error", err))
		}
	}()

	expiresAt, err := plex.TokenExpiry(*pin.AuthToken)
	if err != nil {
		go func() {
			if err := a.plexAuth.SignOut(context.Background(), credentials.ClientId, *pin.AuthToken); err != nil {
				slog.Error("failed to sign out non-jwt plex device", slog.Any("error", err))
			}
		}()
		return nil, connect.NewError(connect.CodeInternal, "plex did not return a jwt").WithCause(err)
	}

	settings, err := a.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeNotFound, "settings not found")
	}

	previousClientId, previousToken := settings.PlexClientId, settings.PlexToken
	settings.PlexClientId = credentials.ClientId
	settings.PlexKeyId = credentials.KeyId
	settings.PlexKeySeed = login.KeySeed
	settings.PlexToken = *pin.AuthToken
	settings.PlexTokenExpiresAt = expiresAt

	if err := a.db.UpdatePlexCredentials(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to update plex auth").WithCause(err)
	}

	if previousToken != "" {
		go func() {
			if err := a.plexAuth.SignOut(context.Background(), previousClientId, previousToken); err != nil {
				slog.Error("failed to sign out previous plex device", slog.Any("error", err))
			}
		}()
	}

	user, err := a.plexApi.GetUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get plex users").WithCause(err)
	}

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
		MaxAge:   30 * 24 * 60 * 60,
	}
	info, _ := connect.CallInfoForServerContext(ctx)
	info.ResponseHeader().Set("Set-Cookie", cookie.String())
	return rsp, nil
}
