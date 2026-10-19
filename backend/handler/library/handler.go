package library

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.LibraryServiceHandler = libraryServiceHandler{}

type libraryServiceHandler struct {
}

func NewHandler() service.LibraryServiceHandler {
	return libraryServiceHandler{}
}

func (l libraryServiceHandler) ListItems(ctx context.Context, req *ListItemsRequest) (*ListItemsResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "not implemented")
}
