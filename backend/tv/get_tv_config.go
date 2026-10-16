package tv

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (t tvServiceHandler) GetTvConfig(ctx context.Context, _ *GetTvConfigRequest) (*GetTvConfigResponse, error) {
	settings, err := t.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "no settings found")
	}

	return &GetTvConfigResponse{
		Config: &TvConfig{
			Address: settings.TvAddress,
			PskSet:  len(settings.TvPsk) > 0,
		},
	}, nil
}
