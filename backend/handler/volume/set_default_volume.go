package volume

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (v volumeServiceHandler) SetDefaultVolume(ctx context.Context, request *SetDefaultVolumeRequest) (*SetDefaultVolumeResponse, error) {
	settings, err := v.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings")
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "setting not found")
	}

	settings.DefaultVolume = int(request.Volume)

	if err := v.db.UpdateDefaultVolume(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to update default_volume").WithCause(err)
	}

	return &SetDefaultVolumeResponse{
		Volume: int32(settings.DefaultVolume),
	}, nil
}
