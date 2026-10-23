package library

import (
	"strconv"

	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.LibraryServiceHandler = libraryServiceHandler{}

type libraryServiceHandler struct {
	db   *database.Database
	plex *plex.ApiClient
}

func NewHandler(db *database.Database, plex *plex.ApiClient) service.LibraryServiceHandler {
	return libraryServiceHandler{
		db:   db,
		plex: plex,
	}
}

func isPlexKey(key string) bool {
	_, err := strconv.ParseUint(key, 10, 64)
	return err == nil
}
