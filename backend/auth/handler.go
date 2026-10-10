package auth

import (
	"context"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
)

var _ service.AuthServiceHandler = authServiceHandler{}

type authServiceHandler struct {
}

func NewHandler() service.AuthServiceHandler {
	return authServiceHandler{}
}

func (a authServiceHandler) StartLogin(ctx context.Context, req *connect.Request[StartLoginRequest]) (*connect.Response[StartLoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (a authServiceHandler) CompleteLogin(ctx context.Context, req *connect.Request[CompleteLoginRequest]) (*connect.Response[CompleteLoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (a authServiceHandler) GetSession(ctx context.Context, req *connect.Request[GetSessionRequest]) (*connect.Response[GetSessionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (a authServiceHandler) Logout(ctx context.Context, req *connect.Request[LogoutRequest]) (*connect.Response[LogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
