package auth

import (
	"context"
	"log"
	"time"

	"connectrpc.com/connect"
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

	users, err := a.plex.GetUsers(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(users.Users) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	user := users.Users[0]

	return &connect.Response[CompleteLoginResponse]{
		Msg: &CompleteLoginResponse{
			Complete: true,
			User: &User{
				Username:  user.Username,
				AvatarUrl: user.AvatarUrl,
			},
		},
	}, nil
}
