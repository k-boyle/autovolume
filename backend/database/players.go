package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type Player struct {
	bun.BaseModel `bun:"table:players"`

	Id       string    `bun:"id,pk"`
	Name     string    `bun:"name,nullzero,notnull"`
	Product  string    `bun:"product,nullzero"`
	LastSeen time.Time `bun:"last_seen"`
}

// todo this should really be a join on settings...
func (db *Database) ReadPlayer(ctx context.Context, id string) (*Player, error) {
	player := new(Player)
	if err := db.db.NewSelect().Model(player).Where("id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return player, nil
}

func (db *Database) InsertPlayer(ctx context.Context, player *Player) error {
	_, err := db.db.NewInsert().Model(player).On("CONFLICT (id) DO UPDATE").Exec(ctx)
	return err
}
