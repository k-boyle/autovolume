package logging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"connectrpc.com/connect/v2"
)

func NewTextHandler() *slog.TextHandler {
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Value.Kind() == slog.KindAny {
				if err, ok := a.Value.Any().(error); ok {
					a.Value = fmtErr(err)
				} else if a.Key == "source" {
					return logSource(a.Value.Any().(*slog.Source))
				}
			}

			return a
		},
	})
}

func NewErrorLogger() connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			err := next(ctx, spec, stream)
			if err == nil || ctx.Err() != nil {
				return nil
			}

			code := connect.CodeOf(err)
			level := slog.LevelError
			switch code {
			case connect.CodeCanceled, connect.CodeInvalidArgument, connect.CodeNotFound,
				connect.CodeAlreadyExists, connect.CodePermissionDenied,
				connect.CodeUnauthenticated, connect.CodeFailedPrecondition:
				level = slog.LevelDebug
			}

			slog.Log(ctx, level, "failed to handle request", slog.String("handler", spec.Procedure), slog.String("code", code.String()), slog.Any("error", err))
			return err
		}
	}
}

func fmtErr(err error) slog.Value {
	msg := []string{err.Error()}
	if cErr, ok := errors.AsType[*connect.Error](err); ok {
		if cErr.Unwrap() != nil {
			msg = append(msg, cErr.Unwrap().Error())
		}
	}

	return slog.StringValue(strings.Join(msg, ": "))
}

func logSource(source *slog.Source) slog.Attr {
	_, after, _ := strings.Cut(source.File, "backend")
	file := slog.String("file", after)
	method := slog.String("method", source.Function)
	line := slog.Int("line", source.Line)
	return slog.GroupAttrs("source", file, method, line)
}
