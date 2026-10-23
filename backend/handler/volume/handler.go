package volume

import (
	"context"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.VolumeServiceHandler = volumeServiceHandler{}

type volumeServiceHandler struct {
	db   *database.Database
	plex *plex.ApiClient
}

func NewHandler(db *database.Database, plex *plex.ApiClient) service.VolumeServiceHandler {
	return volumeServiceHandler{
		db:   db,
		plex: plex,
	}
}

func (v volumeServiceHandler) SetVolumeRule(ctx context.Context, request *SetVolumeRuleRequest) (*SetVolumeRuleResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (v volumeServiceHandler) ClearVolumeRule(ctx context.Context, request *ClearVolumeRuleRequest) (*ClearVolumeRuleResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
