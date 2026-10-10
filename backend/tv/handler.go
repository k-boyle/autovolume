package tv

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.TvServiceHandler = tvServiceHandler{}

type tvServiceHandler struct{}

func NewHandler() service.TvServiceHandler {
	return tvServiceHandler{}
}

func (t tvServiceHandler) GetTvConfig(ctx context.Context, req *connect.Request[GetTvConfigRequest]) (*connect.Response[GetTvConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (t tvServiceHandler) UpdateTvConfig(ctx context.Context, req *connect.Request[UpdateTvConfigRequest]) (*connect.Response[UpdateTvConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (t tvServiceHandler) TestTvConnection(ctx context.Context, req *connect.Request[TestTvConnectionRequest]) (*connect.Response[TestTvConnectionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (t tvServiceHandler) SetTvVolume(ctx context.Context, req *connect.Request[SetTvVolumeRequest]) (*connect.Response[SetTvVolumeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
