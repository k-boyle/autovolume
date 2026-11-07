package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/k-boyle/autovolume/database"
	"github.com/k-boyle/autovolume/plex"
)

type Rotator struct {
	plex   *plex.AuthClient
	db     *database.Database
	ticker *time.Ticker
}

func newRotator(plex *plex.AuthClient, db *database.Database) *Rotator {
	return &Rotator{
		plex: plex,
		db:   db,
	}
}

func (r *Rotator) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Hour)

	if err := r.rotate(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := r.rotate(ctx); err != nil {
				slog.Error("failed to rotate", slog.Any("error", err))
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (r *Rotator) rotate(ctx context.Context) error {
	slog.Info("starting rotate")

	settings, err := r.db.ReadSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return fmt.Errorf("no settings found")
	}
	if settings.PlexToken == "" {
		return nil
	}
	if time.Until(settings.PlexTokenExpiresAt) > 24*time.Hour {
		return nil
	}

	credentials, err := plex.CredentialsFromSeed(settings.PlexClientId, settings.PlexKeyId, settings.PlexKeySeed)
	if err != nil {
		return err
	}

	nonce, err := r.plex.GetNonce(ctx, credentials)
	if err != nil {
		return err
	}

	exchangedToken, err := r.plex.ExchangePlexToken(ctx, credentials, nonce)
	if err != nil {
		return err
	}

	expiresAt, err := plex.TokenExpiry(exchangedToken.AuthToken)
	if err != nil {
		return err
	}

	settings.PlexToken = exchangedToken.AuthToken
	settings.PlexTokenExpiresAt = expiresAt
	if err := r.db.UpdatePlexAuth(ctx, settings); err != nil {
		return err
	}

	slog.Info("rotated auth token")
	return nil
}
