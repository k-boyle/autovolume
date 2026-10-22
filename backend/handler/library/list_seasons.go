package library

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (l libraryServiceHandler) ListSeasons(ctx context.Context, request *ListSeasonsRequest) (*ListSeasonsResponse, error) {
	libraryChildren, err := l.plex.GetLibraryChildren.Execute(ctx, plex.GetLibraryChildrenParams{RatingKey: request.ShowRatingKey})
	if err != nil {
		if se, ok := errors.AsType[*plex.StatusError](err); ok && (se.Code == 400 || se.Code == 404) {
			return nil, connect.NewError(connect.CodeNotFound, "show not found").WithCause(err)
		}
		return nil, connect.NewError(connect.CodeInternal, "failed to get library children").WithCause(err)
	}
	if libraryChildren.ViewGroup != "season" {
		return nil, connect.NewError(connect.CodeNotFound, "not a show")
	}

	var seasons []*MediaItem
	for _, child := range libraryChildren.Metadata {
		thumbUrl := child.ThumbUrl
		if thumbUrl == "" {
			thumbUrl = child.ParentThumbUrl
		}

		// todo season volume
		seasons = append(seasons, &MediaItem{
			RatingKey:     child.RatingKey,
			Type:          MediaType_MEDIA_TYPE_SEASON,
			Title:         child.Title,
			ThumbUrl:      "/api/img" + thumbUrl,
			ShowRatingKey: request.ShowRatingKey,
			ShowTitle:     child.ParentTitle,
			SeasonNumber:  int32(child.Index),
		})
	}

	return &ListSeasonsResponse{
		Seasons: seasons,
	}, nil
}
