package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (a authServiceHandler) GetSession(ctx context.Context, req *connect.Request[GetSessionRequest]) (*connect.Response[GetSessionResponse], error) {
	cookie, err := (&http.Request{Header: req.Header()}).Cookie("av_session")
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}
	session, err := a.db.ReadSession(ctx, cookie.Value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	return &connect.Response[GetSessionResponse]{
		Msg: &GetSessionResponse{
			User: &User{
				Username:  session.Username,
				AvatarUrl: session.AvatarUrl,
			},
		},
	}, nil
}
