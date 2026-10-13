package volume

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.VolumeServiceHandler = volumeServiceHandler{}

type volumeServiceHandler struct {
}

func NewHandler() service.VolumeServiceHandler {
	return volumeServiceHandler{}
}

func (v volumeServiceHandler) SetVolumeRule(ctx context.Context, request *SetVolumeRuleRequest) (*SetVolumeRuleResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (v volumeServiceHandler) ClearVolumeRule(ctx context.Context, request *ClearVolumeRuleRequest) (*ClearVolumeRuleResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (v volumeServiceHandler) ListVolumeRules(ctx context.Context, request *ListVolumeRulesRequest) (*ListVolumeRulesResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (v volumeServiceHandler) GetDefaultVolume(ctx context.Context, request *GetDefaultVolumeRequest) (*GetDefaultVolumeResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (v volumeServiceHandler) SetDefaultVolume(ctx context.Context, request *SetDefaultVolumeRequest) (*SetDefaultVolumeResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
