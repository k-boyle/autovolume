package library

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.LibraryServiceHandler = libraryServiceHandler{}

type libraryServiceHandler struct {
	plex *plex.ApiClient
}

func NewHandler(plex *plex.ApiClient) service.LibraryServiceHandler {
	return libraryServiceHandler{
		plex: plex,
	}
}

func (l libraryServiceHandler) ListItems(ctx context.Context, req *ListItemsRequest) (*ListItemsResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}

func (l libraryServiceHandler) ListSeasons(ctx context.Context, request *ListSeasonsRequest) (*ListSeasonsResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
