package v1

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
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

	resp := GetSearch200JSONResponse{
		HasNext:   hasNext,
		Cursor:    nextCursor,
		ItemCount: len(items),
		Items: make([]struct {
			ChannelId                  uuid.UUID `json:"channel_id"`
			ExternalChannelDisplayName string    `json:"external_channel_display_name"`
			ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
			ExternalChannelId          string    `json:"external_channel_id"`
			ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
			ExternalVideoDescription   string    `json:"external_video_description"`
			ExternalVideoId            string    `json:"external_video_id"`
			ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
			ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
			ExternalVideoTitle         string    `json:"external_video_title"`
			LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
			VideoId                    uuid.UUID `json:"video_id"`
		}, len(items)),
	}

	loc := hutil.TimezoneFromContext(ctx)
	for i, item := range items {
		resp.Items[i].VideoId = item.VideoID
		resp.Items[i].ChannelId = item.ChannelID
		resp.Items[i].ExternalVideoId = item.ExternalVideoID
		resp.Items[i].ExternalChannelId = item.ExternalChannelID
		resp.Items[i].ExternalVideoTitle = item.ExternalVideoTitle
		resp.Items[i].ExternalVideoDescription = item.ExternalVideoDescription
		resp.Items[i].ExternalVideoThumbnailUrl = item.ExternalVideoThumbnailUrl
		resp.Items[i].ExternalChannelDisplayName = item.ExternalChannelDisplayName
		resp.Items[i].ExternalChannelIconUrl = item.ExternalChannelIconUrl
		resp.Items[i].ExternalVideoCreatedAt = item.ExternalVideoCreatedAt.In(loc)
		resp.Items[i].ExternalVideoLengthSeconds = item.ExternalVideoLengthSeconds
	}

	return resp, nil
}

func (h *APIHandler) GetFeed(ctx context.Context, request GetFeedRequestObject) (GetFeedResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	videos, hasNext, err := h.feedService.GetFeed(ctx, userID, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ChannelId                  uuid.UUID `json:"channel_id"`
		ExternalChannelDisplayName string    `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
		ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string    `json:"external_video_title"`
		LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
		VideoId                    uuid.UUID `json:"video_id"`
	}, len(videos))

	loc := hutil.TimezoneFromContext(ctx)
	for i, v := range videos {
		items[i].VideoId = v.VideoId
		items[i].ChannelId = v.ChannelId
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
