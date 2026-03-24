package v1

import (
	"context"

	"github.com/google/uuid"
)

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
