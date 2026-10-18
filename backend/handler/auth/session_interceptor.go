package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
)

var sessionlessProcedures = map[string]bool{
	"/autovolume.v1.AuthService/StartLogin":    true,
	"/autovolume.v1.AuthService/CompleteLogin": true,
}

// /autovolume.v1.AuthService/GetSession
func NewValidateSession(db *database.Database) connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			if _, ok := sessionlessProcedures[spec.Procedure]; ok {
				return next(ctx, spec, stream)
			}

			info, _ := connect.CallInfoForServerContext(ctx)
			c := info.RequestHeader().Get("Cookie")
			cookies, err := http.ParseCookie(c)
			if err != nil {
				return connect.NewError(connect.CodeInvalidArgument, "failed to parse cookies").WithCause(err)
			}
			var sessionId string
			for _, cookie := range cookies {
				if cookie.Name == "av_session" {
					sessionId = cookie.Value
					break
				}
			}
			if sessionId == "" {
				return connect.NewError(connect.CodeUnauthenticated, "no session cookie")
			}

			session, err := db.ReadSession(ctx, sessionId)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return connect.NewError(connect.CodeUnauthenticated, "session not found")
				}
				return connect.NewError(connect.CodeUnknown, "failed to get session").WithCause(err)
			}

			if session.ExpiresAt.Before(time.Now()) {
				go func() {
					err := db.DeleteSession(context.Background(), sessionId)
					if err != nil {
						slog.Error("failed to delete session", slog.Any("error", err))
					}
				}()
				return connect.NewError(connect.CodeUnauthenticated, "session is expired")
			}

			return next(ctx, spec, stream)
		}
	}
}

func isSessionlessProcedure(procedure string) bool {
	return false
}
