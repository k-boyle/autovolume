package auth

import (
	"context"

	"connectrpc.com/connect"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.AuthServiceHandler = authServiceHandler{}

type authServiceHandler struct {
	db   *database.Database
	plex *plex.Client
}

func NewHandler(db *database.Database, plex *plex.Client) service.AuthServiceHandler {
	return authServiceHandler{
		db:   db,
		plex: plex,
	}
}

func (a authServiceHandler) Logout(ctx context.Context, req *connect.Request[LogoutRequest]) (*connect.Response[LogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
