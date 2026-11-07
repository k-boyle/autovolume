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
	"syscall"
	"time"

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
	"github.com/k-boyle/autovolume/logging"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/sony"
	"github.com/spf13/viper"
)

//go:embed web
var web embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.SetDefault(slog.New(logging.NewTextHandler()))

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

	settings, err := ensureSettings(ctx, db)
	if err != nil {
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

	plexWebsocket := plex.NewWebsocket(ctx)
	defer func(websocket *plex.Websocket) {
		err := websocket.Close()
		if err != nil {
			slog.Error("failed to close websocket", slog.Any("error", err))
		}
	}(plexWebsocket)

	if settings.ServerUrl != "" && settings.ServerToken != "" && settings.PlexClientId != "" {
		url, err := url.Parse(settings.ServerUrl)
		if err != nil {
			slog.Error("failed to parse url", slog.Any("error", err))
		}
		if err == nil {
			details := &plex.WebsocketDetails{
				Url:      url,
				Token:    settings.ServerToken,
				ClientId: settings.PlayerId,
			}
			if err := plexWebsocket.ServerChanged(details); err != nil {
				slog.Error("failed to update server details", slog.Any("error", err))
			} else {
				slog.Info("websocket connected", slog.Any("server", url))
			}
		}
	}

	sonyClient := sony.NewClient(http.DefaultClient, func(ctx context.Context) (string, string, error) {
		settings, err := db.ReadSettings(ctx)
		if err != nil {
			return "", "", err
		}

		return settings.TvAddress, settings.TvPsk, nil
	})

	av, err := newAutoVolume(ctx, sonyClient, plexApiClient, plexWebsocket, db)
	if err != nil {
		panic(err)
	}
	defer av.Close()

	mux := http.NewServeMux()

	mux.Handle("/api/", http.StripPrefix("/api", backend(db, plexAuthClient, plexApiClient, plexWebsocket, sonyClient)))
	mux.Handle("/", frontend())

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	srv := &http.Server{
		Addr:      fmt.Sprintf(":%s", config.Port),
		Handler:   mux,
		Protocols: p,
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)

		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown server", slog.Any("error", err))
		}
	}()

	slog.Info(fmt.Sprintf("Listening on %s", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", slog.Any("error", err))
		return
	}
	<-shutdownDone
}

func backend(db *database.Database, authClient *plex.AuthClient, apiClient *plex.ApiClient, websocket *plex.Websocket, sony *sony.Client) http.Handler {
	api := http.NewServeMux()
	api.Handle("GET /img/", thumbnails(db, apiClient))

	server := connect.NewServer(logging.NewErrorLogger(), auth.NewValidateSession(db))
	autovolumev1connect.RegisterAuthServiceHandler(server, auth.NewHandler(db, authClient, apiClient))
	autovolumev1connect.RegisterLibraryServiceHandler(server, library.NewHandler(db, apiClient))
	autovolumev1connect.RegisterNowPlayingServiceHandler(server, playing.NewHandler(db, apiClient, websocket, sony))
	autovolumev1connect.RegisterSetupServiceHandler(server, setup.NewHandler(db, apiClient, websocket))
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

func ensureSettings(ctx context.Context, db *database.Database) (*database.Settings, error) {
	settings, err := db.ReadSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings, err = db.CreateSettings(ctx)
	}

	return settings, err
}
