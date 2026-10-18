package tv

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (t tvServiceHandler) UpdateTvConfig(ctx context.Context, request *UpdateTvConfigRequest) (*UpdateTvConfigResponse, error) {
	settings, err := t.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "no settings found")
	}

	settings.TvAddress = request.Address
	if request.Psk != nil {
		settings.TvPsk = *request.Psk
	}

	if err := t.db.UpdateTV(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to update settings").WithCause(err)
	}

	return &UpdateTvConfigResponse{
		Config: &TvConfig{
			Address: settings.TvAddress,
			PskSet:  len(settings.TvPsk) > 0,
		},
	}, nil
}
