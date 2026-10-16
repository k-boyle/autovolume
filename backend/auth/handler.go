package auth

import (
	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.AuthServiceHandler = authServiceHandler{}

type authServiceHandler struct {
	db   *database.Database
	plex *plex.Client
}

func NewHandler(db *database.Database, plex *plex.Client) service.AuthServiceHandler {
	return authServiceHandler{
		db:   db,
		plex: plex,
	}
}
