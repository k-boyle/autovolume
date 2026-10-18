package auth

import (
	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
)

var _ service.AuthServiceHandler = authServiceHandler{}

type authServiceHandler struct {
	db       *database.Database
	plexAuth *plex.AuthClient
	plexApi  *plex.ApiClient
}

func NewHandler(db *database.Database, authClient *plex.AuthClient, apiClient *plex.ApiClient) service.AuthServiceHandler {
	return authServiceHandler{
		db:       db,
		plexAuth: authClient,
		plexApi:  apiClient,
	}
}
