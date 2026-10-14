package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (a authServiceHandler) GetSession(ctx context.Context, _ *GetSessionRequest) (*GetSessionResponse, error) {
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
		return nil, connect.NewError(connect.CodeUnauthenticated, "no session cookie")
	}

	session, err := a.db.ReadSession(ctx, sessionId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeUnauthenticated, "not a valid session")
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read session").WithCause(err)
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, "session has expired")
	}

	return &GetSessionResponse{
		User: &User{
			Username:  session.Username,
			AvatarUrl: session.AvatarUrl,
		},
	}, nil
}
