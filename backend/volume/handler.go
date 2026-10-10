package volume

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.VolumeServiceHandler = volumeServiceHandler{}

type volumeServiceHandler struct {
}

func NewHandler() service.VolumeServiceHandler {
	return volumeServiceHandler{}
}

func (v volumeServiceHandler) SetVolumeRule(ctx context.Context, req *connect.Request[SetVolumeRuleRequest]) (*connect.Response[SetVolumeRuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (v volumeServiceHandler) ClearVolumeRule(ctx context.Context, req *connect.Request[ClearVolumeRuleRequest]) (*connect.Response[ClearVolumeRuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (v volumeServiceHandler) ListVolumeRules(ctx context.Context, req *connect.Request[ListVolumeRulesRequest]) (*connect.Response[ListVolumeRulesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (v volumeServiceHandler) GetDefaultVolume(ctx context.Context, req *connect.Request[GetDefaultVolumeRequest]) (*connect.Response[GetDefaultVolumeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (v volumeServiceHandler) SetDefaultVolume(ctx context.Context, req *connect.Request[SetDefaultVolumeRequest]) (*connect.Response[SetDefaultVolumeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
