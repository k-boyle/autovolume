package setup

import (
	"context"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.SetupServiceHandler = setupServiceHandler{}

type setupServiceHandler struct {
	db   *database.Database
	plex *plex.Client
}

func NewHandler(db *database.Database, plex *plex.Client) service.SetupServiceHandler {
	return setupServiceHandler{
		db:   db,
		plex: plex,
	}
}

func (s setupServiceHandler) ListServers(ctx context.Context, request *ListServersRequest) (*ListServersResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (s setupServiceHandler) SelectServer(ctx context.Context, request *SelectServerRequest) (*SelectServerResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (s setupServiceHandler) ListPlayers(ctx context.Context, request *ListPlayersRequest) (*ListPlayersResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (s setupServiceHandler) SelectPlayer(ctx context.Context, request *SelectPlayerRequest) (*SelectPlayerResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
