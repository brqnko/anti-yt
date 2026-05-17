package v1

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/feed"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

func (h *APIHandler) GetSearch(ctx context.Context, request GetSearchRequestObject) (GetSearchResponseObject, error) {
	opts := youtube_d.SearchOptions{
		Language:          request.Params.Language,
		PublishedBefore:   request.Params.PublishedBefore,
		PublishedAfter:    request.Params.PublishedAfter,
		RegionCode:        request.Params.RegionCode,
		RelevanceLanguage: request.Params.RelevanceLanguage,
	}
	if request.Params.Order != nil {
		s := string(*request.Params.Order)
		opts.Order = &s
	}
	items, hasNext, nextCursor, err := h.feedService.Search(ctx, request.Params.Query, request.Params.Limit, request.Params.Cursor, opts)
	if err != nil {
		return nil, err
	}

	loc := hutil.TimezoneFromContext(ctx)
	respItems := make([]struct {
		ChannelCustomId            *string          `json:"channel_custom_id,omitempty"`
		ChannelId                  util.Base64UUID  `json:"channel_id"`
		ChannelSubscribersCount    *int             `json:"channel_subscribers_count,omitempty"`
		ExternalChannelDisplayName string           `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string           `json:"external_channel_icon_url"`
		ExternalChannelId          string           `json:"external_channel_id"`
		ExternalVideoCreatedAt     *time.Time       `json:"external_video_created_at,omitempty"`
		ExternalVideoDescription   *string          `json:"external_video_description,omitempty"`
		ExternalVideoId            *string          `json:"external_video_id,omitempty"`
		ExternalVideoLengthSeconds *int             `json:"external_video_length_seconds,omitempty"`
		ExternalVideoThumbnailUrl  *string          `json:"external_video_thumbnail_url,omitempty"`
		ExternalVideoTitle         *string          `json:"external_video_title,omitempty"`
		LastWatchSeconds           *int             `json:"last_watch_seconds,omitempty"`
		Type                       string           `json:"type"`
		VideoId                    *util.Base64UUID `json:"video_id,omitempty"`
	}, len(items))

	for i, item := range items {
		respItems[i].Type = item.Type
		respItems[i].ChannelId = util.Base64UUID(item.ChannelID)
		respItems[i].ExternalChannelId = item.ExternalChannelID
		respItems[i].ExternalChannelDisplayName = item.ExternalChannelDisplayName
		respItems[i].ExternalChannelIconUrl = item.ExternalChannelIconUrl

		if item.Type == "video" {
			videoId := util.Base64UUID(item.VideoID)
			respItems[i].VideoId = &videoId
			respItems[i].ExternalVideoId = &item.ExternalVideoID
			respItems[i].ExternalVideoTitle = &item.ExternalVideoTitle
			respItems[i].ExternalVideoDescription = &item.ExternalVideoDescription
			respItems[i].ExternalVideoThumbnailUrl = &item.ExternalVideoThumbnailUrl
			t := item.ExternalVideoCreatedAt.In(loc)
			respItems[i].ExternalVideoCreatedAt = &t
			respItems[i].ExternalVideoLengthSeconds = &item.ExternalVideoLengthSeconds
		} else {
			respItems[i].ChannelCustomId = &item.ChannelCustomID
			respItems[i].ChannelSubscribersCount = &item.ChannelSubscribersCount
		}
	}

	return GetSearch200JSONResponse{
		HasNext:   hasNext,
		Cursor:    nextCursor,
		ItemCount: len(items),
		Items:     respItems,
	}, nil
}

func (h *APIHandler) GetFeed(ctx context.Context, request GetFeedRequestObject) (GetFeedResponseObject, error) {
	var videos []feed.GetVideoFeedView
	var hasNext bool

	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		videos, hasNext, err = h.feedService.GetLatestVideos(ctx, cursorToUUID(request.Params.Cursor), int32(request.Params.Limit))
	} else {
		videos, hasNext, err = h.feedService.GetFeed(ctx, userID, cursorToUUID(request.Params.Cursor), int32(request.Params.Limit))
	}
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ChannelId                  util.Base64UUID `json:"channel_id"`
		ExternalChannelDisplayName string          `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string          `json:"external_channel_icon_url"`
		ExternalVideoCreatedAt     time.Time       `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int             `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string          `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string          `json:"external_video_title"`
		LastWatchSeconds           *int            `json:"last_watch_seconds,omitempty"`
		VideoId                    util.Base64UUID `json:"video_id"`
	}, len(videos))

	loc := hutil.TimezoneFromContext(ctx)
	for i, v := range videos {
		items[i].VideoId = util.Base64UUID(v.VideoId)
		items[i].ChannelId = util.Base64UUID(v.ChannelId)
		items[i].ExternalChannelIconUrl = v.ExternalChannelIconUrl
		items[i].ExternalChannelDisplayName = v.ExternalChannelDisplayName
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		items[i].ExternalVideoTitle = v.ExternalVideoTitle
		items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt.In(loc)
		items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		items[i].LastWatchSeconds = v.LastWatchSeconds
	}

	return GetFeed200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(items),
		Items:     items,
	}, nil
}
