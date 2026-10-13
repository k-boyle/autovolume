package plex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/k-boyle/autovolume/database"
)

type Rotator struct {
	plex   *Client
	db     *database.Database
	ticker *time.Ticker
}

func NewRotator(plex *Client, db *database.Database) *Rotator {
	return &Rotator{
		plex: plex,
		db:   db,
	}
}

func (r *Rotator) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)

	for {
		select {
		case <-ticker.C:
			if err := r.rotate(ctx); err != nil {
				log.Printf("failed to rotate: %v", err)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (r *Rotator) rotate(ctx context.Context) error {
	log.Println("starting rotate")

	settings, err := r.db.ReadSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return fmt.Errorf("no settings found")
	}
	if settings.PlexToken != "" {
		return nil
	}
	if settings.PlexTokenExpiresAt.Before(time.Now().AddDate(0, 0, -1)) {
		return nil
	}

	nonce, err := r.plex.GetNonce(ctx)
	if err != nil {
		return err
	}

	pin, err := r.plex.ExchangePlexToken(ctx, nonce)
	if err != nil {
		return err
	}

	if pin.AuthToken == nil {
		return fmt.Errorf("no auth token found")
	}

	settings.PlexToken = *pin.AuthToken
	settings.PlexTokenExpiresAt = pin.ExpiresAt
	if err := r.db.UpdatePlexAuth(ctx, settings); err != nil {
		return err
	}

	log.Println("rotated auth token")
	return nil
}
