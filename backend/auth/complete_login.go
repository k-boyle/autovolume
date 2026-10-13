package auth

import (
	"context"
	"crypto/rand"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (a authServiceHandler) CompleteLogin(ctx context.Context, req *connect.Request[CompleteLoginRequest]) (*connect.Response[CompleteLoginResponse], error) {
	login, err := a.db.GetPendingLogin(ctx, req.Msg.LoginId)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	if login == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}
	if login.ExpiresAt.Before(time.Now()) {
		go func() {
			if err := a.db.DeletePendingLogin(ctx, req.Msg.LoginId); err != nil {
				log.Println("failed to delete pending login - ", err)
			}
		}()
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pin, err := a.plex.ExchangePin(ctx, login.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if pin.AuthToken == nil {
		return &connect.Response[CompleteLoginResponse]{
			Msg: &CompleteLoginResponse{},
		}, nil
	}

	go func() {
		if err := a.db.DeletePendingLogin(ctx, req.Msg.LoginId); err != nil {
			log.Println("failed to delete pending login - ", err)
		}
	}()

	settings, err := a.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, nil)
	}

	settings.PlexToken = *pin.AuthToken
	settings.PlexTokenExpiresAt = pin.ExpiresAt

	if err := a.db.UpdatePlexAuth(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	users, err := a.plex.GetUsers(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(users.Users) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, nil)
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
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	rsp := &connect.Response[CompleteLoginResponse]{
		Msg: &CompleteLoginResponse{
			Complete: true,
			User: &User{
				Username:  user.Username,
				AvatarUrl: user.AvatarUrl,
			},
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
	rsp.Header().Add("Set-Cookie", cookie.String())
	return rsp, nil
}
