package auth

import (
	"context"
	"net/http"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (a authServiceHandler) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	info, _ := connect.CallInfoForServerContext(ctx)
	c := info.RequestHeader().Get("Cookie")
	cookies, err := http.ParseCookie(c)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, "failed to parse cookies").WithCause(err)
	}
	var sessionId string
	for _, cookie := range cookies {
		if cookie.Name == "av_session" {
			sessionId = cookie.Value
			break
		}
	}
	if sessionId == "" {
		return &LogoutResponse{}, nil
	}

	if err := a.db.DeleteSession(ctx, sessionId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to delete session").WithCause(err)
	}

	return &LogoutResponse{}, nil
}
