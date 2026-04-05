package v1

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetChannelsChannelId(ctx context.Context, request GetChannelsChannelIdRequestObject) (GetChannelsChannelIdResponseObject, error) {
	detail, err := h.channelService.GetChannelDetail(ctx, request.ChannelId)
	if err != nil {
		return nil, err
	}

	return GetChannelsChannelId200JSONResponse{
		ChannelId:                       detail.ChannelID,
		ExternalChannelCustomId:         detail.CustomID,
		ExternalChannelDisplayName:      detail.DisplayName,
		ExternalChannelDescription:      detail.Description,
		ExternalChannelIconUrl:          detail.IconURL,
		ExternalChannelSubscribersCount: int(detail.SubscribersCount),
	}, nil
}

func (h *APIHandler) GetChannelsChannelIdPlaylists(ctx context.Context, request GetChannelsChannelIdPlaylistsRequestObject) (GetChannelsChannelIdPlaylistsResponseObject, error) {
	playlists, hasNext, err := h.playlistService.GetChannelPlaylists(ctx, request.ChannelId, request.Params.Cursor, int32(request.Params.Limit))
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		PlaylistId           uuid.UUID `json:"playlist_id"`
		PlaylistRegisteredAt time.Time `json:"playlist_registered_at"`
		PlaylistTitle        string    `json:"playlist_title"`
		PlaylistVideoCount   int       `json:"playlist_video_count"`
		TopVideoThumbnailUrl *string   `json:"top_video_thumbnail_url,omitempty"`
	}, len(playlists))

	loc := hutil.TimezoneFromContext(ctx)
	for i, pl := range playlists {
		items[i].PlaylistId = pl.PlaylistId
		items[i].PlaylistTitle = pl.PlaylistTitle
		items[i].PlaylistVideoCount = pl.PlaylistVideoCount
		items[i].PlaylistRegisteredAt = pl.PlaylistRegisteredAt.In(loc)
		items[i].TopVideoThumbnailUrl = pl.TopVideoThumbnailUrl
	}

	return GetChannelsChannelIdPlaylists200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(playlists),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetChannelsChannelIdVideos(ctx context.Context, request GetChannelsChannelIdVideosRequestObject) (GetChannelsChannelIdVideosResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var order string
	if request.Params.Order != nil {
		order = string(*request.Params.Order)
	}

	videos, hasNext, err := h.channelService.GetChannelUploads(ctx, userID, request.ChannelId, request.Params.Cursor, int32(request.Params.Limit), order)
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ExternalVideoCreatedAt     time.Time `json:"external_video_created_at"`
		ExternalVideoLengthSeconds int       `json:"external_video_length_seconds"`
		ExternalVideoThumbnailUrl  string    `json:"external_video_thumbnail_url"`
		ExternalVideoTitle         string    `json:"external_video_title"`
		IsWatched                  bool      `json:"is_watched"`
		LastWatchSeconds           *int      `json:"last_watch_seconds,omitempty"`
		VideoId                    uuid.UUID `json:"video_id"`
	}, len(videos))

	loc := hutil.TimezoneFromContext(ctx)
	for i, v := range videos {
		items[i].VideoId = v.VideoId
		items[i].ExternalVideoThumbnailUrl = v.ExternalVideoThumbnailUrl
		items[i].ExternalVideoTitle = v.ExternalVideoTitle
		items[i].ExternalVideoCreatedAt = v.ExternalVideoCreatedAt.In(loc)
		items[i].ExternalVideoLengthSeconds = v.ExternalVideoLengthSeconds
		items[i].IsWatched = v.IsWatched
		items[i].LastWatchSeconds = v.LastWatchSeconds
	}

	return GetChannelsChannelIdVideos200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(videos),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetFeedChannels(ctx context.Context, request GetFeedChannelsRequestObject) (GetFeedChannelsResponseObject, error) {
	channels, err := h.channelService.GetChannelFeeds(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		CategoryCode               int       `json:"category_code"`
		ChannelId                  uuid.UUID `json:"channel_id"`
		ExternalChannelCusomUrl    string    `json:"external_channel_cusom_url"`
		ExternalChannelDisplayName string    `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
		ValuableDescription        string    `json:"valuable_description"`
	}, len(channels))

	for i, ch := range channels {
		items[i].ChannelId = ch.ChannelId
		items[i].ExternalChannelCusomUrl = ch.ExternalChannelCustomUrl
		items[i].ExternalChannelDisplayName = ch.ExternalChannelDisplayName
		items[i].ExternalChannelIconUrl = ch.ExternalChannelIconUrl
		items[i].CategoryCode = ch.CategoryCode
		items[i].ValuableDescription = ch.ValuableDescription
	}

	return GetFeedChannels200JSONResponse{
		ItemCount: len(items),
		Items:     items,
	}, nil
}

func (h *APIHandler) GetChannelsSubscribed(ctx context.Context, request GetChannelsSubscribedRequestObject) (GetChannelsSubscribedResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	channels, hasNext, err := h.channelService.GetSubscriptions(ctx, userID, int32(request.Params.Limit), request.Params.Cursor)
	if err != nil {
		return nil, err
	}

	items := make([]struct {
		ChannelCustomId            string    `json:"channel_custom_id"`
		ChannelId                  uuid.UUID `json:"channel_id"`
		ChannelSubscribersCount    int       `json:"channel_subscribers_count"`
		ExternalChannelDisplayName string    `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
		ExternalChannelId          string    `json:"external_channel_id"`
	}, len(channels))

	for i, ch := range channels {
		items[i].ChannelId = ch.ChannelId
		items[i].ExternalChannelId = ch.ExternalChannelId
		items[i].ExternalChannelDisplayName = ch.ExternalChannelDisplayName
		items[i].ChannelCustomId = ch.ChannelCustomId
		items[i].ExternalChannelIconUrl = ch.ExternalChannelIconUrl
		items[i].ChannelSubscribersCount = int(ch.ChannelSubscribersCount)
	}

	return GetChannelsSubscribed200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(channels),
		Items:     items,
	}, nil
}

func (h *APIHandler) PostChannelsSubscribe(ctx context.Context, request PostChannelsSubscribeRequestObject) (PostChannelsSubscribeResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	subscribed, err := h.channelService.SubscribeChannel(ctx, userID, request.Body.ChannelId)
	if err != nil {
		return nil, err
	}

	return PostChannelsSubscribe201JSONResponse{
		Body: struct {
			ChannelCreatedAt           time.Time `json:"channel_created_at"`
			ChannelCustomId            string    `json:"channel_custom_id"`
			ChannelDescription         string    `json:"channel_description"`
			ChannelId                  uuid.UUID `json:"channel_id"`
			ChannelSubscribersCount    int       `json:"channel_subscribers_count"`
			ExternalChannelDisplayName string    `json:"external_channel_display_name"`
			ExternalChannelIconUrl     string    `json:"external_channel_icon_url"`
			ExternalChannelId          string    `json:"external_channel_id"`
		}{
			ChannelCreatedAt:           subscribed.Channel.CreatedAt.In(hutil.TimezoneFromContext(ctx)),
			ChannelCustomId:            subscribed.Channel.CustomID,
			ChannelDescription:         subscribed.Channel.Description,
			ChannelId:                  subscribed.ID,
			ChannelSubscribersCount:    int(subscribed.Channel.SubscribersCount),
			ExternalChannelDisplayName: subscribed.Channel.DisplayName,
			ExternalChannelIconUrl:     subscribed.Channel.IconURL,
			ExternalChannelId:          string(subscribed.Channel.ID),
		},
		Headers: PostChannelsSubscribe201ResponseHeaders{
			Location: "/api/v1/channels/" + subscribed.ID.String() + "/subscribe",
		},
	}, nil
}

func (h *APIHandler) DeleteChannelsChannelIdSubscribe(ctx context.Context, request DeleteChannelsChannelIdSubscribeRequestObject) (DeleteChannelsChannelIdSubscribeResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.channelService.UnsubscribeChannel(ctx, userID, request.ChannelId); err != nil {
		return nil, err
	}

	return DeleteChannelsChannelIdSubscribe204Response{}, nil
}
