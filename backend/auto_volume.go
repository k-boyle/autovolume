package main

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/k-boyle/autovolume/database"
	"github.com/k-boyle/autovolume/plex"
	"github.com/k-boyle/autovolume/sony"
)

type AutoVolume struct {
	cancel       context.CancelFunc
	sony         *sony.Client
	api          *plex.ApiClient
	websocket    *plex.Websocket
	subscription *plex.Subscription
	db           *database.Database
	done         chan struct{}
}

func newAutoVolume(ctx context.Context, sony *sony.Client, api *plex.ApiClient, websocket *plex.Websocket, db *database.Database) (*AutoVolume, error) {
	sub, err := websocket.Subscribe()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	av := &AutoVolume{
		cancel:       cancel,
		sony:         sony,
		api:          api,
		websocket:    websocket,
		subscription: sub,
		db:           db,
		done:         make(chan struct{}),
	}

	go av.run(ctx)

	return av, nil
}

func (av *AutoVolume) run(ctx context.Context) {
	defer close(av.done)

	var last *plex.PlaySessionStateNotification
	for notification := range av.subscription.Read() {
		switch notification.State {
		case "playing", "stopped", "paused":
		default:
			continue
		}

		if last != nil && last.RatingKey == notification.RatingKey && last.State == notification.State {
			continue
		}
		last = nil

		settings, err := av.db.ReadSettings(ctx)
		if err != nil {
			slog.Error("failed to read settings", slog.Any("error", err))
			continue
		}

		volume := settings.DefaultVolume
		if notification.State == "playing" {
			metadatas, err := av.api.GetLibraryMetadata(ctx, plex.GetLibraryMetadataParams{RatingKey: notification.RatingKey})
			if err != nil {
				slog.Error("failed to get library metadata", slog.Any("error", err))
				continue
			}
			if len(metadatas.Metadata) == 0 {
				slog.Error("no metadata found")
				continue
			}
			if len(metadatas.Metadata) > 1 {
				slog.Error("multiple metadata found")
				continue
			}

			metadata := metadatas.Metadata[0]
			switch metadata.Type {
			case "episode":
				showRule, err := av.db.ReadVolumeRule(ctx, settings.ServerId, metadata.GrandparentRatingKey)
				if err != nil {
					slog.Error("failed to read rule for show", slog.Any("error", err))
					continue
				}
				if showRule != nil {
					volume = showRule.Volume
				}

				seasonRule, err := av.db.ReadVolumeRule(ctx, settings.ServerId, metadata.ParentRatingKey)
				if err != nil {
					slog.Error("failed to read rule for season", slog.Any("error", err))
					continue
				}
				if seasonRule != nil {
					volume = seasonRule.Volume
				}
			case "movie":
				movieRule, err := av.db.ReadVolumeRule(ctx, settings.ServerId, metadata.RatingKey)
				if err != nil {
					slog.Error("failed to read rule for movie", slog.Any("error", err))
					continue
				}
				if movieRule != nil {
					volume = movieRule.Volume
				}
			default:
				slog.Error("unknown metadata type", slog.Any("type", metadata.Type))
				continue
			}
		}

		if volume == 0 {
			last = notification
			continue
		}

		_, err = av.sony.SetAudioVolume(ctx, sony.SetVolumeInformationParams{
			Volume: strconv.Itoa(volume),
			Output: "",
		})
		if err != nil {
			last = notification
			slog.Error("failed to set audio volume", slog.Any("error", err))
			continue
		}

		last = notification
	}
}

func (av *AutoVolume) Close() {
	av.websocket.Unsubscribe(av.subscription)
	av.cancel()
	<-av.done
}
