package plex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/k-boyle/autovolume/database"
)

type Rotator struct {
	plex   *AuthClient
	db     *database.Database
	ticker *time.Ticker
}

func NewRotator(plex *AuthClient, db *database.Database) *Rotator {
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
			return ctx.Err()
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

	nonce, err := r.plex.GetNonce(ctx)
	if err != nil {
		return err
	}

	exchangedToken, err := r.plex.ExchangePlexToken(ctx, nonce)
	if err != nil {
		return err
	}

	expiresAt, err := tokenExpiry(exchangedToken.AuthToken)
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

func tokenExpiry(token string) (time.Time, error) {
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(token, claims); err != nil {
		return time.Time{}, err
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return time.Time{}, err
	}
	if exp == nil {
		return time.Time{}, errors.New("token has no expiration time")
	}

	return exp.Time, nil
}
