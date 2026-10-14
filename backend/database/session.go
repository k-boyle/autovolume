package database

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel

	Id        string    `bun:"id,pk,notnull"`
	Username  string    `bun:"username,notnull"`
	AvatarUrl string    `bun:"avatar_url,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
}

func (db *Database) InsertSession(ctx context.Context, session *Session) error {
	_, err := db.db.NewInsert().Model(session).Exec(ctx)
	return err
}

func (db *Database) ReadSession(ctx context.Context, id string) (*Session, error) {
	session := new(Session)
	if err := db.db.NewSelect().Model(session).Where("id = ?", id).Scan(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (db *Database) DeleteSession(ctx context.Context, id string) error {
	_, err := db.db.NewDelete().Model(&Session{}).Where("id = ?", id).Exec(ctx)
	return err
}
