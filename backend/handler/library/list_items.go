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

		items = append(items, &MediaItem{
			RatingKey: metadata.RatingKey,
			Type:      libraryType,
			Title:     metadata.Title,
			Year:      int32(metadata.Year),
			ThumbUrl:  "/api/img" + metadata.ThumbUrl,
		})
	}

	return &ListItemsResponse{
		Items: items,
		Total: int32(section.TotalSize),
	}, nil
}
