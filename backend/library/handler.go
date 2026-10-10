package library

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.LibraryServiceHandler = libraryServiceHandler{}

type libraryServiceHandler struct {
}

func NewHandler() service.LibraryServiceHandler {
	return libraryServiceHandler{}
}

func (l libraryServiceHandler) ListLibraries(ctx context.Context, req *connect.Request[ListLibrariesRequest]) (*connect.Response[ListLibrariesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (l libraryServiceHandler) ListItems(ctx context.Context, req *connect.Request[ListItemsRequest]) (*connect.Response[ListItemsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
