package setup

import (
	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.SetupServiceHandler = setupServiceHandler{}

type setupServiceHandler struct {
	db   *database.Database
	plex *plex.Client
}

func NewHandler(db *database.Database, plex *plex.Client) service.SetupServiceHandler {
	return setupServiceHandler{
		db:   db,
		plex: plex,
	}
}
