package playing

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.NowPlayingServiceHandler = nowPlayingServiceHandler{}

type nowPlayingServiceHandler struct {
}

func NewHandler() service.NowPlayingServiceHandler {
	return nowPlayingServiceHandler{}
}

func (n nowPlayingServiceHandler) WatchNowPlaying(ctx context.Context, req *connect.Request[WatchNowPlayingRequest], rsp *connect.ServerStream[WatchNowPlayingResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}
