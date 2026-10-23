package library

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (l libraryServiceHandler) ListItems(ctx context.Context, req *ListItemsRequest) (*ListItemsResponse, error) {
	if !isPlexKey(req.LibraryId) {
		return nil, connect.NewError(connect.CodeInvalidArgument, "invalid library id")
	}

	settings, err := l.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "settings not found")
	}

	section, err := l.plex.GetLibrarySection(ctx, plex.GetLibrarySectionParams{
		Id:    req.LibraryId,
		Title: req.Query,
		Start: int(req.Offset),
		Size:  int(req.Limit),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read library section").WithCause(err)
	}

	var items []*MediaItem
	for _, metadata := range section.Metadata {
		var libraryType = MediaType_MEDIA_TYPE_UNSPECIFIED
		switch metadata.Type {
		case "movie":
			libraryType = MediaType_MEDIA_TYPE_MOVIE
		case "show":
			libraryType = MediaType_MEDIA_TYPE_SHOW
		}

		volume, err := l.db.ReadVolumeRule(ctx, settings.ServerId, metadata.RatingKey)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, "failed to read volume rule").WithCause(err)
		}

		item := &MediaItem{
			RatingKey: metadata.RatingKey,
			Type:      libraryType,
			Title:     metadata.Title,
			Year:      int32(metadata.Year),
			ThumbUrl:  "/api/img" + metadata.ThumbUrl,
		}

		if volume != nil {
			v := int32(volume.Volume)
			item.Volume = &v
		}

		items = append(items, item)
	}

	return &ListItemsResponse{
		Items: items,
		Total: int32(section.TotalSize),
	}, nil
}
