package auth

import (
	"context"
	"crypto/rand"
	"fmt"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (a authServiceHandler) StartLogin(ctx context.Context, _ *StartLoginRequest) (*StartLoginResponse, error) {
	credentials, err := plex.NewCredentials()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to create plex credentials").WithCause(err)
	}

	pin, err := a.plexAuth.CreatePin(ctx, credentials)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, "failed to create pin").WithCause(err)
	}

	pending := &database.PendingLogin{
		Id:        rand.Text(),
		Code:      pin.Code,
		PinId:     pin.Id,
		ClientId:  credentials.ClientId,
		KeyId:     credentials.KeyId,
		KeySeed:   credentials.Key.Seed(),
		CreatedAt: pin.CreatedAt,
		ExpiresAt: pin.ExpiresAt,
	}
	if err := a.db.InsertPendingLogin(ctx, pending); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, "failed to insert pending login").WithCause(err)
	}

	authUrl := fmt.Sprintf(
		"https://app.plex.tv/auth#?clientID=%s&code=%s&context%%5Bdevice%%5D%%5Bproduct%%5D=AutoVolume",
		pin.ClientIdentifier,
		pin.Code,
	)
	return &StartLoginResponse{
		AuthUrl: authUrl,
		LoginId: pending.Id,
	}, nil
}
