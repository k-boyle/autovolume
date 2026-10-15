package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type Settings struct {
	bun.BaseModel `bun:"table:settings"`

	Id                 bool      `bun:"id,pk"`
	PlexClientId       string    `bun:"plex_client_id,nullzero"`
	ServerId           string    `bun:"server_id,nullzero"`
	ServerName         string    `bun:"server_name,nullzero"`
	ServerUrl          string    `bun:"server_url,nullzero"`
	ServerToken        string    `bun:"server_token,nullzero"`
	PlayerId           string    `bun:"player_id,nullzero"`
	TvAddress          string    `bun:"tv_address,nullzero"`
	TvPsk              string    `bun:"tv_psk,nullzero"`
	DefaultVolume      int       `bun:"default_volume"`
	PlexKeySeed        []byte    `bun:"plex_key_seed"`
	PlexToken          string    `bun:"plex_token"`
	PlexTokenExpiresAt time.Time `bun:"plex_token_expires_at"`
}

func (db *Database) ReadSettings(ctx context.Context) (*Settings, error) {
	settings := new(Settings)
	if err := db.db.NewSelect().Model(settings).Where("id = ?", true).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return settings, nil
}

func (db *Database) CreateSettings(ctx context.Context) (*Settings, error) {
	settings := &Settings{
		Id: true,
	}
	pci := new(string)
	_, err := db.db.NewInsert().Model(settings).Returning("plex_client_id").Exec(ctx, pci)
	if err != nil {
		return nil, err
	}
	settings.PlexClientId = *pci
	return settings, nil
}

func (db *Database) UpdatePlexKeySeed(ctx context.Context, settings *Settings) error {
	_, err := db.db.NewUpdate().Model(settings).Column("plex_key_seed").Where("id = ?", settings.Id).Exec(ctx)
	return err
}

func (db *Database) UpdatePlexAuth(ctx context.Context, settings *Settings) error {
	_, err := db.db.NewUpdate().Model(settings).Column("plex_token", "plex_token_expires_at").Where("id = ?", settings.Id).Exec(ctx)
	return err
}

func (db *Database) UpdateServer(ctx context.Context, settings *Settings) error {
	_, err := db.db.NewUpdate().Model(settings).Column("server_id", "server_url", "server_token", "server_name").Where("id = ?", settings.Id).Exec(ctx)
	return err
}
