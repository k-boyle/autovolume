package database

import (
	"context"
	"database/sql"
	"errors"
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

func (db *Database) GetPendingLogin(ctx context.Context, id string) (*PendingLogin, error) {
	login := new(PendingLogin)
	if err := db.db.NewSelect().Model(login).Where("id = ?", id).Scan(ctx, login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return login, nil
}

func (db *Database) DeletePendingLogin(ctx context.Context, id string) error {
	_, err := db.db.NewDelete().Model(&PendingLogin{}).Where("id = ?", id).Exec(ctx)
	return err
}
