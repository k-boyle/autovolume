package setup

import (
	"context"
	"net/url"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s setupServiceHandler) SelectPlayer(ctx context.Context, request *SelectPlayerRequest) (*SelectPlayerResponse, error) {
	settings, err := s.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeNotFound, "settings not found")
	}

	devices, err := s.plex.ListDevices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to list devices")
	}
	var plexPlayer *plex.Device
	for _, device := range devices {
		if device.Id != request.PlayerId {
			continue
		}

		plexPlayer = &device
		break
	}
	if plexPlayer == nil {
		return nil, connect.NewError(connect.CodeNotFound, "device not found")
	}

	url, err := url.Parse(settings.ServerUrl)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "failed to parse server url").WithCause(err)
	}

	details := &plex.WebsocketDetails{
		Url:      url,
		Token:    settings.ServerToken,
		ClientId: plexPlayer.Id,
	}
	if err := s.websocket.ServerChanged(details); err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "failed to connect to server websocket")
	}

	player := &database.Player{
		Id:       plexPlayer.Id,
		Name:     plexPlayer.Name,
		Product:  plexPlayer.Product,
		LastSeen: plexPlayer.LastSeenAt,
	}

	if err := s.db.InsertPlayer(ctx, player); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to insert player").WithCause(err)
	}

	settings.PlayerId = plexPlayer.Id
	if err := s.db.UpdatePlayer(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to update settings").WithCause(err)
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

	return &SelectPlayerResponse{
		Status: &SetupStatus{
			Server: server,
			Player: &Player{
				Id:       player.Id,
				Name:     player.Name,
				Product:  player.Product,
				LastSeen: timestamppb.New(player.LastSeen),
			},
			PlexConnected: settings.ServerUrl != "" && settings.PlexTokenExpiresAt.After(time.Now()),
			TvConfigured:  settings.TvAddress != "",
		},
	}, nil
}
