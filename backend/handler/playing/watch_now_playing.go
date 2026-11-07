package playing

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect/v2"
	"github.com/k-boyle/autovolume/database"
	. "github.com/k-boyle/autovolume/gen/autovolume/v1"
	"github.com/k-boyle/autovolume/gen/autovolume/v1/autovolumev1connect"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/sony"
)

func (n nowPlayingServiceHandler) WatchNowPlaying(ctx context.Context, _ *WatchNowPlayingRequest, stream autovolumev1connect.NowPlayingServiceWatchNowPlayingServerStream) error {
	sub, err := n.websocket.Subscribe()
	if err != nil {
		return connect.NewError(connect.CodeInternal, "failed to subscribe to websocket").WithCause(err)
	}
	defer n.websocket.Unsubscribe(sub)

	var initial *plex.PlaySessionStateNotification
	select {
	case initial = <-sub.Read():
	default:
	}

	if err := n.sendSnapshot(ctx, stream, initial); err != nil {
		return err
	}

	for notification := range sub.Read() {
		if err := n.sendSnapshot(ctx, stream, notification); err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		return connect.NewError(connect.CodeCanceled, "context cancelled").WithCause(ctx.Err())
	}

	return connect.NewError(connect.CodeUnavailable, "lost connection to Plex")
}

func (n nowPlayingServiceHandler) sendSnapshot(ctx context.Context, stream autovolumev1connect.NowPlayingServiceWatchNowPlayingServerStream, notification *plex.PlaySessionStateNotification) error {
	settings, err := n.db.ReadSettings(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, "failed to read settings").WithCause(err)
	}
	if settings == nil {
		return connect.NewError(connect.CodeInternal, "no settings found")
	}

	playbackState := PlaybackState_PLAYBACK_STATE_UNSPECIFIED
	var slogs []slog.Attr
	if notification != nil {
		switch notification.State {
		case "paused":
			playbackState = PlaybackState_PLAYBACK_STATE_PAUSED
		case "playing":
			playbackState = PlaybackState_PLAYBACK_STATE_PLAYING
		}
		slogs = append(slogs, slog.String("state", notification.State))
	}

	var playback *Playback
	if playbackState != PlaybackState_PLAYBACK_STATE_UNSPECIFIED {
		metadatas, err := n.client.GetLibraryMetadata(ctx, plex.GetLibraryMetadataParams{RatingKey: notification.RatingKey})
		if err != nil {
			return connect.NewError(connect.CodeInternal, "failed to get library metadata").WithCause(err)
		}
		if len(metadatas.Metadata) == 0 {
			return connect.NewError(connect.CodeInternal, "no metadata found")
		}
		if len(metadatas.Metadata) > 1 {
			return connect.NewError(connect.CodeInternal, "multiple metadata found")
		}

		metadata := metadatas.Metadata[0]
		switch metadata.Type {
		case "episode":
			playback, err = n.playbackForShow(ctx, settings, metadata)
			if err != nil {
				return connect.NewError(connect.CodeInternal, "failed to watch playback").WithCause(err)
			}
			slogs = append(slogs, slog.String("show", playback.Item.Title), slog.Int("season", int(playback.Season.SeasonNumber)), slog.Int("episode", int(playback.Episode.EpisodeNumber)))
		case "movie":
			playback, err = n.playbackForMovie(ctx, settings, metadata)
			if err != nil {
				return connect.NewError(connect.CodeInternal, "failed to watch playback").WithCause(err)
			}
			slogs = append(slogs, slog.String("film", playback.Item.Title))
		default:
			return connect.NewError(connect.CodeInternal, fmt.Sprintf("unknown playback type %s", metadata.Type))
		}
		playback.State = playbackState
	}

	powerStatus, err := n.sony.GetPowerStatus(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, "failed to get power status").WithCause(err)
	}
	tv := &TvState{}
	switch powerStatus.Status {
	case sony.Standby:
		tv.Power = TvPower_TV_POWER_STANDBY
	case sony.Active:
		tv.Power = TvPower_TV_POWER_ACTIVE
	}

	if tv.Power == TvPower_TV_POWER_ACTIVE {
		volumes, err := n.sony.GetVolumeInformation(ctx, sony.GetVolumeInformationParams{})
		if err != nil {
			return connect.NewError(connect.CodeInternal, "failed to get tv volume").WithCause(err)
		}
		if len(volumes) == 0 {
			return connect.NewError(connect.CodeInternal, "tv returned no volume information")
		}
		volume := int32(volumes[0].Volume)
		tv.Volume = &volume
		tv.Muted = volumes[0].Mute == sony.Muted
	}

	defaultVolume := int32(settings.DefaultVolume)
	response := &WatchNowPlayingResponse{
		Playback:      playback,
		Tv:            tv,
		DefaultVolume: &defaultVolume,
	}
	slog.Info("Plex stream in progress", slog.Any("playback", slogs))

	if err := stream.Send(response); err != nil {
		return connect.NewError(connect.CodeInternal, "failed to send response").WithCause(err)
	}

	return nil
}

func (n nowPlayingServiceHandler) playbackForMovie(ctx context.Context, settings *database.Settings, metadata plex.LibraryMetadata) (*Playback, error) {
	item := &MediaItem{
		RatingKey: metadata.RatingKey,
		Type:      MediaType_MEDIA_TYPE_MOVIE,
		Title:     metadata.Title,
		ThumbUrl:  "/api/img" + metadata.ThumbUrl,
	}

	filmVolumeRule, err := n.db.ReadVolumeRule(ctx, settings.ServerId, metadata.RatingKey)
	if err != nil {
		return nil, err
	}

	if filmVolumeRule != nil {
		volume := int32(filmVolumeRule.Volume)
		item.Volume = &volume
	}

	return &Playback{
		Item: item,
	}, nil
}

func (n nowPlayingServiceHandler) playbackForShow(ctx context.Context, settings *database.Settings, metadata plex.LibraryMetadata) (*Playback, error) {
	show := &MediaItem{
		RatingKey: metadata.GrandparentRatingKey,
		Type:      MediaType_MEDIA_TYPE_SHOW,
		Title:     metadata.GrandparentTitle,
	}
	if metadata.GrandparentThumbUrl != "" {
		show.ThumbUrl = "/api/img" + metadata.GrandparentThumbUrl
	}

	showVolumeRule, err := n.db.ReadVolumeRule(ctx, settings.ServerId, metadata.GrandparentRatingKey)
	if err != nil {
		return nil, err
	}

	if showVolumeRule != nil {
		showVolume := int32(showVolumeRule.Volume)
		show.Volume = &showVolume
	}

	seasonVolumeRule, err := n.db.ReadVolumeRule(ctx, settings.ServerId, metadata.ParentRatingKey)
	if err != nil {
		return nil, err
	}

	season := &MediaItem{
		RatingKey:     metadata.ParentRatingKey,
		Type:          MediaType_MEDIA_TYPE_SEASON,
		Title:         metadata.ParentTitle,
		ShowRatingKey: metadata.GrandparentRatingKey,
		ShowTitle:     metadata.GrandparentTitle,
		SeasonNumber:  int32(metadata.ParentIndex),
	}

	if seasonVolumeRule != nil {
		seasonVolume := int32(seasonVolumeRule.Volume)
		season.Volume = &seasonVolume
	}

	if metadata.ParentThumbUrl != "" {
		season.ThumbUrl = "/api/img" + metadata.ParentThumbUrl
	}

	episode := &Episode{
		SeasonNumber:  int32(metadata.ParentIndex),
		EpisodeNumber: int32(metadata.Index),
		Title:         metadata.Title,
	}

	return &Playback{
		Item:    show,
		Season:  season,
		Episode: episode,
	}, nil
}
