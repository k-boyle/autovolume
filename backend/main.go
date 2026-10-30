package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/k-boyle/autovolume/database"
	"github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/handler/auth"
	"github.com/k-boyle/autovolume/handler/library"
	"github.com/k-boyle/autovolume/handler/playing"
	"github.com/k-boyle/autovolume/handler/setup"
	"github.com/k-boyle/autovolume/handler/tv"
	"github.com/k-boyle/autovolume/handler/volume"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/sony"
	"github.com/spf13/viper"
)

//go:embed web
var web embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	textHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
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
	slog.SetDefault(slog.New(textHandler))

	v := viper.New()
	config, err := LoadConfig(v)
	if err != nil {
		slog.Error("failed to read config", slog.Any("error", err))
		return
	}
	slog.Info("Config loaded")

	db, err := database.New(config.DatabaseConnection)
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		return
	}
	defer func(db *database.Database) {
		err := db.Close()
		if err != nil {
			slog.Error("failed to close database connection", slog.Any("error", err))
		}
	}(db)

	if err := db.Migrate(); err != nil {
		slog.Error("failed to run migrations", slog.Any("error", err))
		return
	}
	slog.Info("Migrations run")

	if err := ensureSettings(ctx, db); err != nil {
		slog.Error("failed to get settings", slog.Any("error", err))
		return
	}
	plexAuthClient := plex.NewAuthClient(http.DefaultClient)

	plexApiClient := plex.NewApiClient(http.DefaultClient,
		func(ctx context.Context) (*url.URL, string, string, error) {
			settings, err := db.ReadSettings(ctx)
			if err != nil {
				return nil, "", "", err
			}

			return plex.Url, settings.PlexToken, settings.PlexClientId, nil
		},
		func(ctx context.Context) (*url.URL, string, string, error) {
			settings, err := db.ReadSettings(ctx)
			if err != nil {
				return nil, "", "", err
			}
			u, err := url.Parse(settings.ServerUrl)
			if err != nil {
				return nil, "", "", err
			}

			return u, settings.ServerToken, settings.PlexClientId, nil
		},
	)

	rotator := plex.NewRotator(plexAuthClient, db)
	go func() {
		err := rotator.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	sonyClient := sony.NewClient(http.DefaultClient, func(ctx context.Context) (string, string, error) {
		settings, err := db.ReadSettings(ctx)
		if err != nil {
			return "", "", err
		}

		return settings.TvAddress, settings.TvPsk, nil
	})

	mux := http.NewServeMux()

	mux.Handle("/api/", http.StripPrefix("/api", backend(db, plexAuthClient, plexApiClient, sonyClient)))
	mux.Handle("/", frontend())

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	srv := &http.Server{
		Addr:      fmt.Sprintf(":%s", config.Port),
		Handler:   mux,
		Protocols: p,
	}

	slog.Info(fmt.Sprintf("Listening on %s", srv.Addr))
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
	}
}

func backend(db *database.Database, authClient *plex.AuthClient, apiClient *plex.ApiClient, sony *sony.Client) http.Handler {
	api := http.NewServeMux()
	api.Handle("GET /img/", thumbnails(db, apiClient))

	server := connect.NewServer(NewErrorLogger(), auth.NewValidateSession(db))
	autovolumev1connect.RegisterAuthServiceHandler(server, auth.NewHandler(db, authClient, apiClient))
	autovolumev1connect.RegisterLibraryServiceHandler(server, library.NewHandler(db, apiClient))
	autovolumev1connect.RegisterNowPlayingServiceHandler(server, playing.NewHandler(db, apiClient, sony))
	autovolumev1connect.RegisterSetupServiceHandler(server, setup.NewHandler(db, apiClient))
	autovolumev1connect.RegisterTvServiceHandler(server, tv.NewHandler(db, sony))
	autovolumev1connect.RegisterVolumeServiceHandler(server, volume.NewHandler(db, apiClient))
	connecthttp.Mount(api, server)

	return api
}

func frontend() http.Handler {
	dist, err := fs.Sub(web, "web/dist")
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		isAsset := strings.HasPrefix(name, "assets/")

		switch {
		case isFile(dist, name):
			if isAsset {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFileFS(w, r, dist, name)

		case isAsset:
			http.NotFound(w, r)

		default:
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFileFS(w, r, dist, "index.html")
		}
	})
}

func thumbnails(db *database.Database, apiClient *plex.ApiClient) http.Handler {
	validSession := auth.IsValidSession(db)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")
		valid, err := validSession(r.Context(), cookie)
		if err != nil {
			slog.Error("failed to validate session", slog.Any("error", err))
			http.Error(w, "failed to validate session", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		thumb := strings.TrimPrefix(r.URL.Path, "/img")
		if !strings.HasPrefix(thumb, "/library/metadata/") {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		res, err := apiClient.GetThumb(r.Context(), thumb, q.Get("w"), q.Get("h"))
		if err != nil {
			slog.Error("failed to get thumbnail", slog.Any("error", err))
			http.Error(w, "failed to fetch image", http.StatusBadGateway)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				slog.Error("failed to close response body", slog.Any("error", err))
			}
		}(res.Body)
		if res.StatusCode != http.StatusOK {
			w.WriteHeader(res.StatusCode)
			return
		}
		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.Header().Set("Cache-Control", "private, max-age=2592000")
		if _, err := io.Copy(w, res.Body); err != nil {
			http.Error(w, "failed to write image", http.StatusBadGateway)
			return
		}
	})
}

func isFile(f fs.FS, name string) bool {
	info, err := fs.Stat(f, name)
	return err == nil && !info.IsDir()
}

func ensureSettings(ctx context.Context, db *database.Database) error {
	settings, err := db.ReadSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		_, err = db.CreateSettings(ctx)
	}

	return err
}

func NewErrorLogger() connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			if err := next(ctx, spec, stream); err != nil {
				if !errors.Is(err, context.Canceled) {
					slog.Error("failed to handle request", slog.String("handler", spec.Procedure), slog.Any("error", err))
				}
				return err
			}

			return nil
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
