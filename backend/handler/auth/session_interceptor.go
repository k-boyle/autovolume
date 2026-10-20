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

func NewValidateSession(db *database.Database) connect.ServerInterceptor {
	validSession := IsValidSession(db)
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			if _, ok := sessionlessProcedures[spec.Procedure]; ok {
				return next(ctx, spec, stream)
			}

			info, _ := connect.CallInfoForServerContext(ctx)
			h := info.RequestHeader()
			valid, err := validSession(ctx, h.Get("Cookie"))
			if err != nil {
				return connect.NewError(connect.CodeInternal, "failed to validate session").WithCause(err)
			}
			if !valid {
				return connect.NewError(connect.CodeUnauthenticated, "invalid session")
			}

			return next(ctx, spec, stream)
		}
	}
}

func IsValidSession(db *database.Database) func(context.Context, string) (bool, error) {
	return func(ctx context.Context, cookies string) (bool, error) {
		c, err := http.ParseCookie(cookies)
		if err != nil {
			return false, err
		}
		var sessionId string
		for _, cookie := range c {
			if cookie.Name == "av_session" {
				sessionId = cookie.Value
				break
			}
		}

		if sessionId == "" {
			return false, nil
		}

		session, err := db.ReadSession(ctx, sessionId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return false, err
		}

		if session.ExpiresAt.Before(time.Now()) {
			go func() {
				err := db.DeleteSession(context.Background(), sessionId)
				if err != nil {
					slog.Error("failed to delete session", slog.Any("error", err))
				}
			}()
			return false, nil
		}

		return true, nil
	}
}
