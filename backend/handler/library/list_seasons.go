package library

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (l libraryServiceHandler) ListSeasons(ctx context.Context, request *ListSeasonsRequest) (*ListSeasonsResponse, error) {
	if !isPlexKey(request.ShowRatingKey) {
		return nil, connect.NewError(connect.CodeInvalidArgument, "invalid show rating key")
	}

	settings, err := l.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "settings not found")
	}

	libraryChildren, err := l.plex.GetLibraryChildren(ctx, plex.GetLibraryChildrenParams{RatingKey: request.ShowRatingKey})
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

		volume, err := l.db.ReadVolumeRule(ctx, settings.ServerId, child.RatingKey)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, "failed to read volume rule").WithCause(err)
		}

		season := &MediaItem{
			RatingKey:     child.RatingKey,
			Type:          MediaType_MEDIA_TYPE_SEASON,
			Title:         child.Title,
			ThumbUrl:      "/api/img" + thumbUrl,
			ShowRatingKey: request.ShowRatingKey,
			ShowTitle:     child.ParentTitle,
			SeasonNumber:  int32(child.Index),
		}

		if volume != nil {
			v := int32(volume.Volume)
			season.Volume = &v
		}

		seasons = append(seasons, season)
	}

	return &ListSeasonsResponse{
		Seasons: seasons,
	}, nil
}
