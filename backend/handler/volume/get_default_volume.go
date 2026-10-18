package volume

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (v volumeServiceHandler) GetDefaultVolume(ctx context.Context, _ *GetDefaultVolumeRequest) (*GetDefaultVolumeResponse, error) {
	settings, err := v.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeNotFound, "setting not found")
	}

	if settings.DefaultVolume > 0 {
		volume := int32(settings.DefaultVolume)
		return &GetDefaultVolumeResponse{
			Volume: &volume,
		}, nil
	}

	return &GetDefaultVolumeResponse{}, nil
}
