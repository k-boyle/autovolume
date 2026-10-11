package auth

import (
	"context"
	"crypto/rand"
	"fmt"

	"connectrpc.com/connect"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (a authServiceHandler) StartLogin(ctx context.Context, req *connect.Request[StartLoginRequest]) (*connect.Response[StartLoginResponse], error) {
	pin, err := a.plex.CreatePin(ctx)
	if err != nil {
		return nil, err
	}

	pending := &database.PendingLogin{
		Id:        rand.Text(),
		Code:      pin.Code,
		PinId:     pin.Id,
		CreatedAt: pin.CreatedAt,
		ExpiresAt: pin.ExpiresAt,
	}
	if err := a.db.InsertPendingLogin(ctx, pending); err != nil {
		return nil, err
	}

	authUrl := fmt.Sprintf(
		"%s/auth#?clientID=%s&code=%s&context%%5Bdevice%%5D%%5Bproduct%%5D=AutoVolume",
		plex.AppUrl,
		pin.ClientIdentifier,
		pin.Code,
	)
	return &connect.Response[StartLoginResponse]{
		Msg: &StartLoginResponse{
			AuthUrl: authUrl,
			LoginId: pending.Id,
		},
	}, nil
}
