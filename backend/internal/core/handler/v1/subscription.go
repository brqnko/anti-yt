package v1

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/util"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *APIHandler) GetSubscriptions(c context.Context, request GetSubscriptionsRequestObject) (GetSubscriptionsResponseObject, error) {
	channels, hasNext, err := h.channelService.GetSubscriptions(c, request.Params.Limit, request.Params.Cursor)
	if err != nil {
		if errors.Is(err, channel.ErrInvalidSubscriptionLimit) {
			return GetSubscriptions400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		util.LogError(c, err)
		return GetSubscriptions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	items := make([]struct {
		ChannelCreatedAt           time.Time          `json:"channel_created_at"`
		ChannelCustomId            string             `json:"channel_custom_id"`
		ChannelDescription         string             `json:"channel_description"`
		ChannelId                  openapi_types.UUID `json:"channel_id"`
		ChannelSubscribersCount    int                `json:"channel_subscribers_count"`
		CreatedAt                  time.Time          `json:"created_at"`
		ExternalChannelDisplayName string             `json:"external_channel_display_name"`
		ExternalChannelIconUrl     string             `json:"external_channel_icon_url"`
		ExternalChannelId          string             `json:"external_channel_id"`
		SubscriptionId             openapi_types.UUID `json:"subscription_id"`
	}, len(channels))

	for i, ch := range channels {
		items[i].SubscriptionId = ch.SubscriptionId
		items[i].ChannelId = ch.ChannelId
		items[i].CreatedAt = ch.CreatedAt
		items[i].ExternalChannelId = string(*ch.ExternalChannelInfo.Id)
		items[i].ExternalChannelDisplayName = ch.ExternalChannelInfo.DisplayName
		items[i].ChannelCustomId = string(*ch.ExternalChannelInfo.CustomId)
		items[i].ChannelDescription = ch.ExternalChannelInfo.Description
		items[i].ExternalChannelIconUrl = ch.ExternalChannelInfo.IconUrl
		items[i].ChannelSubscribersCount = ch.ExternalChannelInfo.SubscribersCount
		items[i].ChannelCreatedAt = ch.ExternalChannelInfo.CreatedAt
	}

	return GetSubscriptions200JSONResponse{
		HasNext:   hasNext,
		ItemCount: len(channels),
		Items:     items,
	}, nil
}

func (h *APIHandler) PostSubscriptions(c context.Context, request PostSubscriptionsRequestObject) (PostSubscriptionsResponseObject, error) {
	subscribed, err := h.channelService.SubscribeChannel(c, request.Body.ChannelId)
	if err != nil {
		if errors.Is(err, util.ErrInvalidYouTubeURL) ||
			errors.Is(err, util.ErrInvalidChannelID) ||
			errors.Is(err, util.ErrInvalidChannelHandle) {
			return PostSubscriptions400JSONResponse{BadRequestJSONResponse{
				Detail: err.Error(),
				Title:  "Bad Request",
			}}, nil
		}

		util.LogError(c, err)
		return PostSubscriptions500JSONResponse{
			InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
				Detail: internalErrorDetail,
				Title:  internalErrorTitle,
			},
		}, nil
	}

	return PostSubscriptions201JSONResponse{
		Body: struct {
			ChannelCreatedAt           time.Time          `json:"channel_created_at"`
			ChannelCustomId            string             `json:"channel_custom_id"`
			ChannelDescription         string             `json:"channel_description"`
			ChannelId                  openapi_types.UUID `json:"channel_id"`
			ChannelSubscribersCount    int                `json:"channel_subscribers_count"`
			CreatedAt                  time.Time          `json:"created_at"`
			ExternalChannelDisplayName string             `json:"external_channel_display_name"`
			ExternalChannelIconUrl     string             `json:"external_channel_icon_url"`
			ExternalChannelId          string             `json:"external_channel_id"`
			SubscriptionId             openapi_types.UUID `json:"subscription_id"`
		}{
			ChannelCreatedAt:           subscribed.ExternalChannelInfo.CreatedAt,
			ChannelCustomId:            string(*subscribed.ExternalChannelInfo.CustomId),
			ChannelDescription:         subscribed.ExternalChannelInfo.Description,
			ChannelId:                  subscribed.ChannelId,
			ChannelSubscribersCount:    subscribed.ExternalChannelInfo.SubscribersCount,
			CreatedAt:                  subscribed.CreatedAt,
			ExternalChannelDisplayName: subscribed.ExternalChannelInfo.DisplayName,
			ExternalChannelIconUrl:     subscribed.ExternalChannelInfo.IconUrl,
			ExternalChannelId:          string(*subscribed.ExternalChannelInfo.Id),
			SubscriptionId:             subscribed.SubscriptionId,
		},
		Headers: PostSubscriptions201ResponseHeaders{
			Location: "/api/v1/subscriptions/" + subscribed.SubscriptionId.String(),
		},
	}, nil
}

func (h *APIHandler) DeleteSubscriptionsSubscriptionId(c context.Context, request DeleteSubscriptionsSubscriptionIdRequestObject) (DeleteSubscriptionsSubscriptionIdResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
