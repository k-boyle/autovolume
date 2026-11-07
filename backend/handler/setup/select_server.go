package setup

import (
	"context"
	"log/slog"
	"net/url"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s setupServiceHandler) SelectServer(ctx context.Context, request *SelectServerRequest) (*SelectServerResponse, error) {
	settings, err := s.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeNotFound, "settings not found")
	}

	devices, err := s.plex.ListDevices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to list devices").WithCause(err)
	}
	var plexDevice *plex.Device
	for _, device := range devices {
		if device.Id != request.ServerId {
			continue
		}

		plexDevice = &device
		break
	}
	if plexDevice == nil {
		return nil, connect.NewError(connect.CodeNotFound, "device not found")
	}

	var connected *plex.Connection
	for _, connection := range plexDevice.Connections {
		_, err := s.plex.GetIdentity(ctx, plex.Override{
			Url: connection.Uri,
		})
		if err == nil {
			connected = connection
			break
		}

		slog.Error("failed to connect to server", slog.Any("error", err))
	}
	if connected == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "failed to connect to server")
	}

	url, err := url.Parse(connected.Uri)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "failed to parse server url").WithCause(err)
	}

	details := &plex.WebsocketDetails{
		Url:      url,
		Token:    plexDevice.AccessToken,
		ClientId: settings.PlayerId,
	}
	if err := s.websocket.ServerChanged(details); err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "failed to connect to server websocket")
	}

	settings.ServerId = request.ServerId
	settings.ServerUrl = connected.Uri
	settings.ServerName = plexDevice.Name
	settings.ServerToken = plexDevice.AccessToken

	if err := s.db.UpdateServer(ctx, settings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to update settings").WithCause(err)
	}

	server := &Server{
		Id:     settings.ServerId,
		Name:   settings.ServerName,
		Owned:  true,
		Online: plexDevice.Presence,
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

	return &SelectServerResponse{
		Status: &SetupStatus{
			Server:        server,
			Player:        player,
			PlexConnected: true,
			TvConfigured:  settings.TvAddress != "",
		},
	}, nil
}
