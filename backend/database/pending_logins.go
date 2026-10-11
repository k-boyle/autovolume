package database

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type PendingLogin struct {
	bun.BaseModel `bun:"table:pending_logins"`

	Id    string `bun:"id,pk,nullzero,notnull"`
	Code  string `bun:"code,nullzero,notnull"`
	PinId int    `bun:"pin_id,nullzero,notnull"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull"`
	ExpiresAt time.Time `bun:"expires_at,nullzero,notnull"`
}

func (db *Database) InsertPendingLogin(ctx context.Context, pending *PendingLogin) error {
	_, err := db.db.NewInsert().Model(pending).Exec(ctx)
	return err
}
