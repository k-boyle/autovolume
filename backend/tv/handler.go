package tv

import (
	"context"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/sony"
)

var _ service.TvServiceHandler = tvServiceHandler{}

type tvServiceHandler struct {
	db   *database.Database
	sony *sony.Client
}

func NewHandler(db *database.Database, sony *sony.Client) service.TvServiceHandler {
	return tvServiceHandler{
		db:   db,
		sony: sony,
	}
}

func (t tvServiceHandler) SetTvVolume(ctx context.Context, request *SetTvVolumeRequest) (*SetTvVolumeResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
