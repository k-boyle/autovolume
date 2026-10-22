package library

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
)

func (l libraryServiceHandler) ListLibraries(ctx context.Context, _ *ListLibrariesRequest) (*ListLibrariesResponse, error) {
	librarySections, err := l.plex.GetAllLibrarySections(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get library items").WithCause(err)
	}

	var library []*Library
	for _, directory := range librarySections.Directory {
		var libraryType = MediaType_MEDIA_TYPE_UNSPECIFIED
		switch directory.Type {
		case "movie":
			libraryType = MediaType_MEDIA_TYPE_MOVIE
		case "show":
			libraryType = MediaType_MEDIA_TYPE_SHOW
		}

		library = append(library, &Library{
			Id:    directory.Key,
			Title: directory.Title,
			Type:  libraryType,
		})
	}

	return &ListLibrariesResponse{
		Libraries: library,
	}, nil
}
