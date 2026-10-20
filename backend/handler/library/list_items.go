package library

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (l libraryServiceHandler) ListItems(ctx context.Context, req *ListItemsRequest) (*ListItemsResponse, error) {
	section, err := l.plex.GetLibrarySection.Execute(ctx, plex.GetLibrarySectionPathParams{Id: req.LibraryId})
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
			ThumbUrl:  "/img" + metadata.ThumbUrl,
		})
	}

	return &ListItemsResponse{
		Items: items,
		Total: int32(len(section.Metadata)),
	}, nil
}
