package volume

import (
	"context"
	"strings"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/plex"
)

func (v volumeServiceHandler) ListVolumeRules(ctx context.Context, _ *ListVolumeRulesRequest) (*ListVolumeRulesResponse, error) {
	rules, err := v.db.ListVolumeRules(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to list volume rules").WithCause(err)
	}

	if len(rules) == 0 {
		return &ListVolumeRulesResponse{}, nil
	}

	var ratingKeys []string
	rulesByRatingKey := map[string]*database.VolumeRule{}
	for _, rule := range rules {
		rulesByRatingKey[rule.RatingKey] = rule
		ratingKeys = append(ratingKeys, rule.RatingKey)
	}

	ratingKeyQuery := strings.Join(ratingKeys, ",")
	metadatas, err := v.plex.GetLibraryMetadata(ctx, plex.GetLibraryMetadataParams{RatingKey: ratingKeyQuery})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "failed to read library metadata").WithCause(err)
	}

	var mediaItems []*MediaItem
	for _, metadata := range metadatas.Metadata {
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

		rule := rulesByRatingKey[metadata.RatingKey]
		volume := int32(rule.Volume)
		item := &MediaItem{
			RatingKey: metadata.RatingKey,
			Type:      libraryType,
			Title:     metadata.Title,
			Year:      int32(metadata.Year),
			ThumbUrl:  "/api/img" + thumbUrl,
			Volume:    &volume,
		}

		if libraryType == MediaType_MEDIA_TYPE_SEASON {
			item.ShowRatingKey = metadata.ParentRatingKey
			item.ShowTitle = metadata.ParentTitle
			item.SeasonNumber = int32(metadata.Index)
		}
		mediaItems = append(mediaItems, item)
	}

	return &ListVolumeRulesResponse{
		Items: mediaItems,
	}, nil
}
