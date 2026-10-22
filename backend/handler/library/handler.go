package library

import (
	"strconv"

	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.LibraryServiceHandler = libraryServiceHandler{}

type libraryServiceHandler struct {
	plex *plex.ApiClient
}

func NewHandler(plex *plex.ApiClient) service.LibraryServiceHandler {
	return libraryServiceHandler{
		plex: plex,
	}
}

func isPlexKey(key string) bool {
	_, err := strconv.ParseUint(key, 10, 64)
	return err == nil
}
