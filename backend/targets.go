package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/k-boyle/autovolume/database"
	"github.com/k-boyle/autovolume/httpx"
	"github.com/k-boyle/autovolume/plex"
)

func plexTvTarget(db *database.Database) httpx.Resolver {
	return func(ctx context.Context) (*httpx.Target, error) {
		settings, err := readSettings(ctx, db)
		if err != nil {
			return nil, err
		}

		return &httpx.Target{
			Base: plex.Url,
			Header: http.Header{
				"X-Plex-Token":             {settings.PlexToken},
				"X-Plex-Client-Identifier": {settings.PlexClientId},
			},
		}, nil
	}
}

func pmsTarget(db *database.Database) httpx.Resolver {
	return func(ctx context.Context) (*httpx.Target, error) {
		settings, err := readSettings(ctx, db)
		if err != nil {
			return nil, err
		}

		base, err := parseBase(settings.ServerUrl)
		if err != nil {
			return nil, err
		}

		return &httpx.Target{
			Base: base,
			Header: http.Header{
				"X-Plex-Token":             {settings.ServerToken},
				"X-Plex-Client-Identifier": {settings.PlexClientId},
			},
		}, nil
	}
}

func tvTarget(db *database.Database) httpx.Resolver {
	return func(ctx context.Context) (*httpx.Target, error) {
		settings, err := readSettings(ctx, db)
		if err != nil {
			return nil, err
		}

		base, err := parseBase(settings.TvAddress)
		if err != nil {
			return nil, err
		}

		return &httpx.Target{
			Base: base,
			Header: http.Header{
				"X-Auth-PSK": {settings.TvPsk},
			},
		}, nil
	}
}

func readSettings(ctx context.Context, db *database.Database) (*database.Settings, error) {
	settings, err := db.ReadSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.New("no settings found")
	}

	return settings, nil
}

func parseBase(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}

	return url.Parse(raw)
}
