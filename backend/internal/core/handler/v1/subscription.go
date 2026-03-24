package v1

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func (h *APIHandler) GetChannelsSubscribed(ctx context.Context, request GetChannelsSubscribedRequestObject) (GetChannelsSubscribedResponseObject, error) {
	userID, err := hutil.UserIDFromContext(ctx)
	if err != nil {
		hutil.LogError(ctx, err)
		return GetChannelsSubscribed500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	channels, hasNext, err := h.channelService.GetSubscriptions(ctx, userID, int32(request.Params.Limit), request.Params.Cursor)
	if err != nil {
		if errors.Is(err, channel.ErrInvalidSubscriptionLimit) {
			return GetChannelsSubscribed400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return GetChannelsSubscribed500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
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
		hutil.LogError(ctx, err)
		return PostChannelsSubscribe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	subscribed, err := h.channelService.SubscribeChannel(ctx, userID, request.Body.ChannelId)
	if err != nil {
		if errors.Is(err, channel.ErrInvalidYouTubeURL) ||
			errors.Is(err, channel.ErrInvalidChannelID) ||
			errors.Is(err, channel.ErrInvalidChannelHandle) {
			return PostChannelsSubscribe400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return PostChannelsSubscribe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
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
			ChannelCreatedAt:           subscribed.Channel.CreatedAt,
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
		hutil.LogError(ctx, err)
		return DeleteChannelsChannelIdSubscribe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	if err := h.channelService.UnsubscribeChannel(ctx, userID, request.ChannelId); err != nil {
		if errors.Is(err, channel.ErrSubscriptionNotFound) {
			return DeleteChannelsChannelIdSubscribe404JSONResponse{NotFoundJSONResponse{
				Detail: "subscription not found",
				Title:  "Not Found",
			}}, nil
		}

		hutil.LogError(ctx, err)
		return DeleteChannelsChannelIdSubscribe500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return DeleteChannelsChannelIdSubscribe204Response{}, nil
}
