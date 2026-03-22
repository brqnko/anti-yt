package youtube_d

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"strings"

	"github.com/mmcdole/gofeed"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var iso8601DurationRe = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

var (
	ErrInvalidChannelID              = errors.New("invalid channel id")
	ErrChannelNotFound               = errors.New("requested channel not found")
	ErrInvalidChannelSnippetResponse = errors.New("channel snippet not found in response")

	ErrVideoIDsTooMuch = errors.New("video ids are too much")
	ErrInvalidVideoID  = errors.New("invalid video id")
)

type ChannelDetail struct {
	ID                string
	DisplayName       string
	CustomID          string
	Description       string
	IconURL           string
	SubscribersCount  int
	UploadsPlaylistID string
	CreatedAt         time.Time
}

type RSSFeedVideo struct {
	VideoID      string
	Title        string
	ThumbnailURL string
	Description  string
	CreatedAt    time.Time
}

type VideoDetail struct {
	ID            string
	ChannelID     string
	Title         string
	Description   string
	ThumbnailURL  string
	LengthSeconds int
	CreatedAt     time.Time
}

type YouTubeAPIService interface {
	FetchChannelDetail(ctx context.Context, channelIDs []string) (map[string]*ChannelDetail, error)
	FetchRSSFeed(ctx context.Context, channelID string) ([]RSSFeedVideo, error)
	FetchVideoDetail(ctx context.Context, videoIDs []string) (map[string]VideoDetail, error)
	FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (videoIDs []string, nextPageToken string, err error)
}

var _ YouTubeAPIService = (*youTubeAPIServiceImpl)(nil)

type youTubeAPIServiceImpl struct {
	ytClient   *youtube.Service
	feedParser *gofeed.Parser
}

// NOTE: 削除された動画などはAPIから返ってこず、また順番も保証されないのでmapで返す
func (s *youTubeAPIServiceImpl) FetchVideoDetail(ctx context.Context, videoIDs []string) (map[string]VideoDetail, error) {
	if len(videoIDs) == 0 {
		return map[string]VideoDetail{}, nil
	}
	if len(videoIDs) > 50 {
		return nil, ErrVideoIDsTooMuch
	}
	for _, videoID := range videoIDs {
		if len(videoID) != 11 {
			return nil, ErrInvalidVideoID
		}
	}

	res, err := s.ytClient.Videos.List([]string{"snippet", "contentDetails"}).
		Id(videoIDs...).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to Videos.List: %w", err)
	}

	videos := make(map[string]VideoDetail, len(res.Items))
	for _, item := range res.Items {
		lengthSeconds := 0
		if item.ContentDetails != nil {
			matches := iso8601DurationRe.FindStringSubmatch(item.ContentDetails.Duration)
			if matches != nil {
				hours, _ := strconv.Atoi(matches[1])
				minutes, _ := strconv.Atoi(matches[2])
				seconds, _ := strconv.Atoi(matches[3])
				lengthSeconds = hours*3600 + minutes*60 + seconds
			}
		}

		detail := VideoDetail{
			ID:            item.Id,
			LengthSeconds: lengthSeconds,
		}
		if item.Snippet != nil {
			detail.ChannelID = item.Snippet.ChannelId
			detail.Title = item.Snippet.Title
			detail.Description = item.Snippet.Description
			if item.Snippet.Thumbnails != nil {
				if item.Snippet.Thumbnails.Medium != nil {
					detail.ThumbnailURL = item.Snippet.Thumbnails.Medium.Url
				} else if item.Snippet.Thumbnails.Default != nil {
					detail.ThumbnailURL = item.Snippet.Thumbnails.Default.Url
				}
			}
			if t, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt); err == nil {
				detail.CreatedAt = t
			}
		}
		videos[item.Id] = detail
	}

	return videos, nil
}

func (s *youTubeAPIServiceImpl) FetchChannelDetail(ctx context.Context, channelIDs []string) (map[string]*ChannelDetail, error) {
	if len(channelIDs) == 0 {
		return map[string]*ChannelDetail{}, nil
	}

	parts := []string{"snippet", "statistics", "contentDetails"}
	result := make(map[string]*ChannelDetail, len(channelIDs))

	// ハンドル(@...)はForHandleで個別にリクエスト、UC IDはまとめてリクエスト
	var ucIDs []string
	for _, id := range channelIDs {
		if strings.HasPrefix(id, "@") && len([]rune(id)) > 3 {
			res, err := s.ytClient.Channels.List(parts).ForHandle(id).Context(ctx).Do()
			if err != nil {
				return nil, fmt.Errorf("failed to Channels.List(handle=%s): %w", id, err)
			}
			if len(res.Items) == 0 {
				continue
			}
			found := res.Items[0]
			if found.Snippet == nil {
				return nil, ErrInvalidChannelSnippetResponse
			}
			createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse: %w", err)
			}
			iconURL := ""
			if found.Snippet.Thumbnails != nil {
				if found.Snippet.Thumbnails.Medium != nil {
					iconURL = found.Snippet.Thumbnails.Medium.Url
				} else if found.Snippet.Thumbnails.Default != nil {
					iconURL = found.Snippet.Thumbnails.Default.Url
				}
			}
			if iconURL == "" {
				slog.Warn("icon url not foud when fetch channel info", "channel id", id)
			}
			subscribersCount := 0
			if found.Statistics != nil {
				subscribersCount = int(found.Statistics.SubscriberCount)
			}
			uploadsPlaylistID := ""
			if found.ContentDetails != nil && found.ContentDetails.RelatedPlaylists != nil {
				uploadsPlaylistID = found.ContentDetails.RelatedPlaylists.Uploads
			}
			result[id] = &ChannelDetail{
				ID:                found.Id,
				DisplayName:       found.Snippet.Title,
				CustomID:          found.Snippet.CustomUrl,
				Description:       found.Snippet.Description,
				IconURL:           iconURL,
				SubscribersCount:  subscribersCount,
				UploadsPlaylistID: uploadsPlaylistID,
				CreatedAt:         createdAt,
			}
		} else if strings.HasPrefix(id, "UC") && len(id) == 24 {
			ucIDs = append(ucIDs, id)
		} else {
			return nil, ErrInvalidChannelID
		}
	}

	if len(ucIDs) > 0 {
		res, err := s.ytClient.Channels.List(parts).Id(ucIDs...).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to Channels.List: %w", err)
		}
		for _, found := range res.Items {
			if found.Snippet == nil {
				return nil, ErrInvalidChannelSnippetResponse
			}
			createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse: %w", err)
			}
			iconURL := ""
			if found.Snippet.Thumbnails != nil {
				if found.Snippet.Thumbnails.Medium != nil {
					iconURL = found.Snippet.Thumbnails.Medium.Url
				} else if found.Snippet.Thumbnails.Default != nil {
					iconURL = found.Snippet.Thumbnails.Default.Url
				}
			}
			if iconURL == "" {
				slog.Warn("icon url not foud when fetch channel info", "channel id", found.Id)
			}
			subscribersCount := 0
			if found.Statistics != nil {
				subscribersCount = int(found.Statistics.SubscriberCount)
			}
			uploadsPlaylistID := ""
			if found.ContentDetails != nil && found.ContentDetails.RelatedPlaylists != nil {
				uploadsPlaylistID = found.ContentDetails.RelatedPlaylists.Uploads
			}
			result[found.Id] = &ChannelDetail{
				ID:                found.Id,
				DisplayName:       found.Snippet.Title,
				CustomID:          found.Snippet.CustomUrl,
				Description:       found.Snippet.Description,
				IconURL:           iconURL,
				SubscribersCount:  subscribersCount,
				UploadsPlaylistID: uploadsPlaylistID,
				CreatedAt:         createdAt,
			}
		}
	}

	return result, nil
}

func (s *youTubeAPIServiceImpl) FetchRSSFeed(ctx context.Context, channelID string) ([]RSSFeedVideo, error) {
	if !strings.HasPrefix(channelID, "UC") || len(channelID) != 24 {
		return nil, ErrInvalidChannelID
	}

	feed, err := s.feedParser.ParseURLWithContext(
		fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID),
		ctx,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to ParseURLWithContext(channel=%s): %w", channelID, err)
	}

	videos := make([]RSSFeedVideo, len(feed.Items))
	for i, item := range feed.Items {
		videos[i] = RSSFeedVideo{
			Title: item.Title,
		}

		if item.PublishedParsed != nil {
			videos[i].CreatedAt = *item.PublishedParsed
		}

		// yt:videoId
		if ytExt, ok := item.Extensions["yt"]; ok {
			if videoIDExt, ok := ytExt["videoId"]; ok && len(videoIDExt) > 0 {
				videos[i].VideoID = videoIDExt[0].Value
			}
		}

		// media:group 内の thumbnail, description, statistics
		if mediaExt, ok := item.Extensions["media"]; ok {
			if groups, ok := mediaExt["group"]; ok && len(groups) > 0 {
				group := groups[0]

				if thumbnails, ok := group.Children["thumbnail"]; ok && len(thumbnails) > 0 {
					videos[i].ThumbnailURL = thumbnails[0].Attrs["url"]
				}

				if descriptions, ok := group.Children["description"]; ok && len(descriptions) > 0 {
					videos[i].Description = descriptions[0].Value
				}
			}
		}
	}

	return videos, nil
}

func (s *youTubeAPIServiceImpl) FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) ([]string, string, error) {
	call := s.ytClient.PlaylistItems.List([]string{"contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(50).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to PlaylistItems.List(playlist=%s): %w", playlistID, err)
	}

	videoIDs := make([]string, 0, len(res.Items))
	for _, item := range res.Items {
		if item.ContentDetails != nil && item.ContentDetails.VideoId != "" {
			videoIDs = append(videoIDs, item.ContentDetails.VideoId)
		}
	}

	return videoIDs, res.NextPageToken, nil
}

func NewYouTubeAPIServiceImpl(ctx context.Context, apiKey string) (YouTubeAPIService, error) {
	ytClient, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to NewService: %w", err)
	}

	feedParser := gofeed.NewParser()
	feedParser.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"

	return &youTubeAPIServiceImpl{
		ytClient:   ytClient,
		feedParser: feedParser,
	}, nil
}
