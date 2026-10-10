package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/k-boyle/autovolume/auth"
	"github.com/k-boyle/autovolume/database"
	"github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/library"
	"github.com/k-boyle/autovolume/playing"
	"github.com/k-boyle/autovolume/setup"
	"github.com/k-boyle/autovolume/tv"
	"github.com/k-boyle/autovolume/volume"
	"github.com/spf13/viper"
)

//go:embed web
var web embed.FS

func main() {
	v := viper.New()
	config, err := LoadConfig(v)
	if err != nil {
		log.Fatal("failed to read config", err)
	}
	log.Println("Config loaded")

	db, err := database.New(config.DatabaseConnection)
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	defer func(db *database.Database) {
		err := db.Close()
		if err != nil {
			log.Fatal("failed to close database connection", err)
		}
	}(db)

	if err := db.Migrate(); err != nil {
		log.Fatal("failed to run migrations", err)
	}
	log.Println("Migrations run")

	mux := http.NewServeMux()

	mux.Handle("/api/", http.StripPrefix("/api", backend()))
	mux.Handle("/", frontend())

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	srv := &http.Server{
		Addr:      fmt.Sprintf(":%s", config.Port),
		Handler:   mux,
		Protocols: p,
	}

	log.Println("Listening on " + srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("failed to start server", err)
	}
}

func backend() http.Handler {
	api := http.NewServeMux()
	pattern, handler := autovolumev1connect.NewAuthServiceHandler(auth.NewHandler())
	api.Handle(pattern, handler)

	pattern, handler = autovolumev1connect.NewLibraryServiceHandler(library.NewHandler())
	api.Handle(pattern, handler)

	pattern, handler = autovolumev1connect.NewNowPlayingServiceHandler(playing.NewHandler())
	api.Handle(pattern, handler)

	pattern, handler = autovolumev1connect.NewSetupServiceHandler(setup.NewHandler())
	api.Handle(pattern, handler)

	pattern, handler = autovolumev1connect.NewTvServiceHandler(tv.NewHandler())
	api.Handle(pattern, handler)

	pattern, handler = autovolumev1connect.NewVolumeServiceHandler(volume.NewHandler())
	api.Handle(pattern, handler)
	return api
}

func frontend() http.Handler {
	dist, err := fs.Sub(web, "web/dist")
	if err != nil {
		log.Fatal(err)
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
