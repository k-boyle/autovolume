package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type VolumeRule struct {
	bun.BaseModel `bun:"table:volume_rules"`

	ServerId  string    `bun:"server_id,pk"`
	RatingKey string    `bun:"rating_key,pk"`
	Volume    int       `bun:"volume"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

func (db *Database) ListVolumeRules(ctx context.Context) ([]*VolumeRule, error) {
	var rules []*VolumeRule
	if err := db.db.NewSelect().Model(new(VolumeRule)).Scan(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (db *Database) InsertVolumeRule(ctx context.Context, rule *VolumeRule) error {
	_, err := db.db.NewInsert().Model(rule).On("CONFLICT (server_id,rating_key) DO UPDATE SET updated_at = now(), volume = ?", rule.Volume).Exec(ctx)
	return err
}

func (db *Database) ReadVolumeRule(ctx context.Context, serverId, ratingKey string) (*VolumeRule, error) {
	rule := new(VolumeRule)
	if err := db.db.NewSelect().Model(rule).Where("server_id = ? AND rating_key = ?", serverId, ratingKey).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rule, nil
}

func (db *Database) DeleteVolumeRule(ctx context.Context, serverId, ratingKey string) error {
	_, err := db.db.NewDelete().Model(new(VolumeRule)).Where("server_id = ? AND rating_key = ?", serverId, ratingKey).Exec(ctx)
	return err
}
