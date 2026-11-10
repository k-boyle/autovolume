package tv

import (
	"context"
	"strconv"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/sony"
)

func (t tvServiceHandler) SetTvVolume(ctx context.Context, request *SetTvVolumeRequest) (*SetTvVolumeResponse, error) {
	_, err := t.sony.SetAudioVolume(ctx, sony.SetVolumeInformationParams{
		Volume: strconv.Itoa(int(request.Volume)),
		Output: "",
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to set the volume").WithCause(err)
	}

	return &SetTvVolumeResponse{}, nil
}
