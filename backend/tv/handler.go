package tv

import (
	"context"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.TvServiceHandler = tvServiceHandler{}

type tvServiceHandler struct {
	db *database.Database
}

func NewHandler(db *database.Database) service.TvServiceHandler {
	return tvServiceHandler{
		db: db,
	}
}

func (t tvServiceHandler) UpdateTvConfig(ctx context.Context, request *UpdateTvConfigRequest) (*UpdateTvConfigResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (t tvServiceHandler) TestTvConnection(ctx context.Context, request *TestTvConnectionRequest) (*TestTvConnectionResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (t tvServiceHandler) SetTvVolume(ctx context.Context, request *SetTvVolumeRequest) (*SetTvVolumeResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
