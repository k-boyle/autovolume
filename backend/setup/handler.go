package setup

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.SetupServiceHandler = setupServiceHandler{}

type setupServiceHandler struct{}

func NewHandler() service.SetupServiceHandler {
	return setupServiceHandler{}
}

func (s setupServiceHandler) ListServers(ctx context.Context, req *connect.Request[ListServersRequest]) (*connect.Response[ListServersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s setupServiceHandler) SelectServer(ctx context.Context, req *connect.Request[SelectServerRequest]) (*connect.Response[SelectServerResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s setupServiceHandler) ListPlayers(ctx context.Context, req *connect.Request[ListPlayersRequest]) (*connect.Response[ListPlayersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s setupServiceHandler) SelectPlayer(ctx context.Context, req *connect.Request[SelectPlayerRequest]) (*connect.Response[SelectPlayerResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s setupServiceHandler) GetSetupStatus(ctx context.Context, req *connect.Request[GetSetupStatusRequest]) (*connect.Response[GetSetupStatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
