package playing

import (
	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/sony"
)

var _ service.NowPlayingServiceHandler = nowPlayingServiceHandler{}

type nowPlayingServiceHandler struct {
	db     *database.Database
	client *plex.ApiClient
	sony   *sony.Client
}

func NewHandler(db *database.Database, client *plex.ApiClient, sony *sony.Client) service.NowPlayingServiceHandler {
	return nowPlayingServiceHandler{
		db:     db,
		client: client,
		sony:   sony,
	}
}
