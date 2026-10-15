package setup

import (
	"context"
	"slices"
	"strings"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s setupServiceHandler) ListPlayers(ctx context.Context, _ *ListPlayersRequest) (*ListPlayersResponse, error) {
	devices, err := s.plex.ListDevices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to list devices")
	}
	var plexPlayers []*plex.Device
	for _, device := range devices {
		if !slices.Contains(strings.Split(device.Provides, ","), "controller") {
			continue
		}

		plexPlayers = append(plexPlayers, device)
	}

	var players []*Player
	for _, plexPlayer := range plexPlayers {
		players = append(players, &Player{
			Id:       plexPlayer.Id,
			Name:     plexPlayer.Name,
			Product:  plexPlayer.Product,
			LastSeen: timestamppb.New(plexPlayer.LastSeenAt),
		})
	}

	return &ListPlayersResponse{
		Players: players,
	}, nil
}
