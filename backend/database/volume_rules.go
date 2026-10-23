package database

import (
	"context"
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
