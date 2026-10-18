package tv

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/sony"
)

func (t tvServiceHandler) TestTvConnection(ctx context.Context, request *TestTvConnectionRequest) (*TestTvConnectionResponse, error) {
	info, err := t.sony.GetInterfaceInformation.Execute(ctx, nil, sony.Override{
		Url: request.Address,
		Psk: request.Psk,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get tv information").WithCause(err)
	}

	volumes, err := t.sony.GetVolumeInformation.Execute(ctx, &sony.GetVolumeInformationParams{}, sony.Override{
		Url: request.Address,
		Psk: request.Psk,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get tv volume").WithCause(err)
	}
	volume := (*volumes)[0]

	power, err := t.sony.GetPowerStatus.Execute(ctx, nil, sony.Override{
		Url: request.Address,
		Psk: request.Psk,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get tv power").WithCause(err)
	}

	var tvPower int
	switch power.Status {
	case sony.Standby:
		tvPower = 2
	case sony.Active:
		tvPower = 3
	}

	v := int32(volume.Volume)
	return &TestTvConnectionResponse{
		Model: info.ModelName,
		State: &TvState{
			Power:  TvPower(tvPower),
			Volume: &v,
			Muted:  volume.Mute == sony.NotMuted,
		},
	}, nil
}
