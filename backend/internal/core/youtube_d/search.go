package youtube_d

import (
	"context"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
)

type SearchItemType string

const (
	SearchItemTypeVideo   SearchItemType = "video"
	SearchItemTypeChannel SearchItemType = "channel"
)

type SearchItem struct {
	Type      SearchItemType
	VideoID   VideoID
	ChannelID ChannelID
}

func (s *clientImpl) SearchIDs(ctx context.Context, query string, pageToken string, opts SearchOptions) (_ []SearchItem, _ string, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).SearchIDs")
	defer s.markIfQuotaExceeded(&err)

	if err := s.checkQuota(); err != nil {
		return nil, "", err
	}

	call := s.ytClient.Search.List([]string{"id"}).
		Q(query).
		MaxResults(50).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if opts.Language != nil && *opts.Language != "" {
		call = call.RelevanceLanguage(*opts.Language)
	}
	if opts.Order != nil && *opts.Order != "" {
		call = call.Order(*opts.Order)
	}
	if opts.PublishedBefore != nil {
		call = call.PublishedBefore(opts.PublishedBefore.Format(time.RFC3339))
	}
	if opts.PublishedAfter != nil {
		call = call.PublishedAfter(opts.PublishedAfter.Format(time.RFC3339))
	}
	if opts.RegionCode != nil && *opts.RegionCode != "" {
		call = call.RegionCode(*opts.RegionCode)
	}
	if opts.RelevanceLanguage != nil && *opts.RelevanceLanguage != "" {
		call = call.RelevanceLanguage(*opts.RelevanceLanguage)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	items := make([]SearchItem, 0, len(res.Items))
	for _, item := range res.Items {
		if item.Id == nil {
			continue
		}
		switch {
		case item.Id.VideoId != "":
			videoID, err := NewVideoID(item.Id.VideoId)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse video id(search ids)", slog.Any("error", err))
				continue
			}
			items = append(items, SearchItem{Type: SearchItemTypeVideo, VideoID: videoID})
		case item.Id.ChannelId != "":
			channelID, err := NewChannelID(item.Id.ChannelId)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse channel id(search ids)", slog.Any("error", err))
				continue
			}
			items = append(items, SearchItem{Type: SearchItemTypeChannel, ChannelID: channelID})
		}
	}

	return items, res.NextPageToken, nil
}
