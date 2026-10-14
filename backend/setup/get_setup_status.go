package setup

import (
	"context"
	"time"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s setupServiceHandler) GetSetupStatus(ctx context.Context, _ *GetSetupStatusRequest) (*GetSetupStatusResponse, error) {
	settings, err := s.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "settings not found")
	}

	var server *Server
	if settings.ServerId != "" {
		server = &Server{
			Id:     settings.ServerId,
			Name:   settings.ServerName,
			Owned:  true,
			Online: settings.ServerUrl != "",
		}
	}

	p, err := s.db.ReadPlayer(ctx, settings.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read player").WithCause(err)
	}

	var player *Player
	if p != nil {
		player = &Player{
			Id:       p.Id,
			Name:     p.Name,
			LastSeen: timestamppb.New(p.LastSeen),
			Product:  p.Product,
		}
	}

	return &GetSetupStatusResponse{
		Status: &SetupStatus{
			Server:        server,
			Player:        player,
			PlexConnected: settings.ServerUrl != "" && settings.PlexTokenExpiresAt.After(time.Now()),
			TvConfigured:  settings.TvAddress != "" && settings.TvPsk != "",
		},
	}, nil
}
