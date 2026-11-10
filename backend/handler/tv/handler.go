package tv

import (
	"github.com/k-boyle/autovolume/database"
	service "github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/sony"
)

var _ service.TvServiceHandler = tvServiceHandler{}

type tvServiceHandler struct {
	db   *database.Database
	sony *sony.Client
}

func NewHandler(db *database.Database, sony *sony.Client) service.TvServiceHandler {
	return tvServiceHandler{
		db:   db,
		sony: sony,
	}
}
