package volume

import (
	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.VolumeServiceHandler = volumeServiceHandler{}

type volumeServiceHandler struct {
	db   *database.Database
	plex *plex.ApiClient
}

func NewHandler(db *database.Database, plex *plex.ApiClient) service.VolumeServiceHandler {
	return volumeServiceHandler{
		db:   db,
		plex: plex,
	}
}
