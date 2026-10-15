package setup

import (
	"context"
	"slices"
	"strings"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (s setupServiceHandler) ListServers(ctx context.Context, _ *ListServersRequest) (*ListServersResponse, error) {
	devices, err := s.plex.ListDevices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to list plex devices").WithCause(err)
	}

	var servers []*Server
	for _, device := range devices {
		if !slices.Contains(strings.Split(device.Provides, ","), "server") {
			continue
		}

		servers = append(servers, &Server{
			Id:     device.Id,
			Name:   device.Name,
			Owned:  true,
			Online: device.Presence,
		})
	}

	return &ListServersResponse{
		Servers: servers,
	}, nil
}
