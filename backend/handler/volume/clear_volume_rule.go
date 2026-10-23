package volume

import (
	"context"

	"connectrpc.com/connect/v2"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (v volumeServiceHandler) ClearVolumeRule(ctx context.Context, request *ClearVolumeRuleRequest) (*ClearVolumeRuleResponse, error) {
	settings, err := v.db.ReadSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return nil, connect.NewError(connect.CodeInternal, "no settings found")
	}

	metadatas, err := v.plex.GetLibraryMetadata(ctx, plex.GetLibraryMetadataParams{RatingKey: request.RatingKey})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to get metadata")
	}
	if len(metadatas.Metadata) == 0 {
		return nil, connect.NewError(connect.CodeInternal, "metadata not found")
	}
	if len(metadatas.Metadata) > 1 {
		return nil, connect.NewError(connect.CodeInternal, "multiple metadata found")
	}

	metadata := metadatas.Metadata[0]
	var libraryType = MediaType_MEDIA_TYPE_UNSPECIFIED
	switch metadata.Type {
	case "movie":
		libraryType = MediaType_MEDIA_TYPE_MOVIE
	case "show":
		libraryType = MediaType_MEDIA_TYPE_SHOW
	case "season":
		libraryType = MediaType_MEDIA_TYPE_SEASON
	}

	thumbUrl := metadata.ThumbUrl
	if thumbUrl == "" {
		thumbUrl = metadata.ParentThumbUrl
	}

	item := &MediaItem{
		RatingKey: metadata.RatingKey,
		Type:      libraryType,
		Title:     metadata.Title,
		Year:      int32(metadata.Year),
		ThumbUrl:  "/api/img" + thumbUrl,
	}

	if err := v.db.DeleteVolumeRule(ctx, settings.ServerId, request.RatingKey); err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to clear rule")
	}

	return &ClearVolumeRuleResponse{
		Item: item,
	}, nil
}
