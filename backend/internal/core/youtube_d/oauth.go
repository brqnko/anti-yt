package youtube_d

import (
	"context"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"google.golang.org/api/youtube/v3"
)

type OAuthClient struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time

	ytClient *youtube.Service
}

func (c *OAuthClient) FetchAllSubscriptions(ctx context.Context) (_ []Channel, err error) {
	defer util.Wrap(&err, "youtube_d.(*OAuthClient).FetchAllSubscriptions")

	var channels []Channel
	pageToken := ""
	for {
		call := c.ytClient.Subscriptions.List([]string{"snippet"}).Mine(true).MaxResults(50).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, item := range resp.Items {
			channels = append(channels, Channel{
				ID:          ChannelID(item.Snippet.ResourceId.ChannelId),
				DisplayName: item.Snippet.Title,
				IconURL:     item.Snippet.Thumbnails.Default.Url,
			})
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return channels, nil
}

func (c *OAuthClient) FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "youtube_d.(*OAuthClient).FetchPlaylistVideoIDs")

	call := c.ytClient.PlaylistItems.List([]string{"contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(50).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	videoIDs := make([]VideoID, 0, len(res.Items))
	for _, item := range res.Items {
		if item.ContentDetails == nil || item.ContentDetails.VideoId == "" {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.ContentDetails is nil or VideoId is empty(fetch playlist video ids with oauth)")
			continue
		}
		videoID, err := NewVideoID(item.ContentDetails.VideoId)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video id(fetch playlist video ids with oauth)", slog.Any("error", err))
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}