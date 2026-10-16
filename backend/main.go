package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/k-boyle/autovolume/auth"
	"github.com/k-boyle/autovolume/database"
	"github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/library"
	"github.com/k-boyle/autovolume/playing"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/setup"
	"github.com/k-boyle/autovolume/tv"
	"github.com/k-boyle/autovolume/volume"
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

	settings, err := getSettings(ctx, db)
	if err != nil {
		slog.Error("failed to get settings", slog.Any("error", err))
		return
	}
	pub, prv := getKeys(settings)

	client, err := plex.New(http.DefaultClient, settings.PlexClientId, pub, prv, func(ctx context.Context) (string, error) {
		settings, err := db.ReadSettings(ctx)
		if err != nil {
			return "", err
		}

		return settings.PlexToken, nil
	})
	if err != nil {
		slog.Error("failed to create plex client", slog.Any("error", err))
		return
	}

	rotator := plex.NewRotator(client, db)
	go rotator.Start(ctx)

	mux := http.NewServeMux()

	mux.Handle("/api/", http.StripPrefix("/api", backend(db, client)))
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

func backend(db *database.Database, plex *plex.Client) http.Handler {
	api := http.NewServeMux()
	server := connect.NewServer(NewErrorLogger(), auth.NewValidateSession(db))
	autovolumev1connect.RegisterAuthServiceHandler(server, auth.NewHandler(db, plex))
	autovolumev1connect.RegisterLibraryServiceHandler(server, library.NewHandler())
	autovolumev1connect.RegisterNowPlayingServiceHandler(server, playing.NewHandler())
	autovolumev1connect.RegisterSetupServiceHandler(server, setup.NewHandler(db, plex))
	autovolumev1connect.RegisterTvServiceHandler(server, tv.NewHandler())
	autovolumev1connect.RegisterVolumeServiceHandler(server, volume.NewHandler(db))
	connecthttp.Mount(api, server)
	return api
}

func frontend() http.Handler {
	dist, err := fs.Sub(web, "web/dist")
	if err != nil {
		slog.Error("failed to open embedded web directory", slog.Any("error", err))
		// :^)
		log.Fatal()
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

func isFile(f fs.FS, name string) bool {
	info, err := fs.Stat(f, name)
	return err == nil && !info.IsDir()
}

func getSettings(ctx context.Context, db *database.Database) (*database.Settings, error) {
	settings, err := db.ReadSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings, err = db.CreateSettings(ctx)
		if err != nil {
			return nil, err
		}
	}

	if len(settings.PlexKeySeed) == 0 {
		_, prv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}

		settings.PlexKeySeed = prv.Seed()
		if err := db.UpdatePlexKeySeed(ctx, settings); err != nil {
			return nil, err
		}

		return settings, nil
	}

	return settings, nil
}

func getKeys(settings *database.Settings) (ed25519.PublicKey, ed25519.PrivateKey) {
	prv := ed25519.NewKeyFromSeed(settings.PlexKeySeed)
	return prv.Public().(ed25519.PublicKey), prv
}

func NewErrorLogger() connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			if err := next(ctx, spec, stream); err != nil {
				slog.Error("failed to handle request", slog.Any("error", err))
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
