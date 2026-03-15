package youtube_d

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var iso8601DurationRe = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

var (
	ErrInvalidChannelId              = errors.New("invalid channel id")
	ErrChannelNotFound               = errors.New("requested channel not found")
	ErrInvalidChannelSnippetResponse = errors.New("channel snippet not found in response")

	ErrVideoIdsTooMuch = errors.New("video ids are too much")
	ErrInvalidVideoId  = errors.New("invalid video id")
)

type ChannelInfo struct {
	Id               string
	DisplayName      string
	CustomId         string
	Description      string
	IconUrl          string
	SubscribersCount int
	CreatedAt        time.Time
}

type RSSFeedVideo struct {
	VideoId      string
	Title        string
	ThumbnailUrl string
	Description  string
	CreatedAt    time.Time
}

type VideoInfo struct {
	Id            string
	LengthSeconds int
}

type YouTubeAPIService interface {
	FetchChannelInfo(ctx context.Context, channelId string) (*ChannelInfo, error)
	FetchRSSFeed(ctx context.Context, channelId string) ([]RSSFeedVideo, error)
	FetchVideoInfo(ctx context.Context, videoIds []string) (map[string]VideoInfo, error)
}

var _ YouTubeAPIService = (*youTubeAPIServiceImpl)(nil)

type youTubeAPIServiceImpl struct {
	ytClient   *youtube.Service
	feedParser *gofeed.Parser
}

// NOTE: 削除された動画などはAPIから返ってこず、また順番も保証されないのでmapで返す
func (s *youTubeAPIServiceImpl) FetchVideoInfo(ctx context.Context, videoIds []string) (map[string]VideoInfo, error) {
	if len(videoIds) == 0 {
		return map[string]VideoInfo{}, nil
	}
	if len(videoIds) > 50 {
		return nil, ErrVideoIdsTooMuch
	}
	for _, videoId := range videoIds {
		if len(videoId) != 11 {
			return nil, ErrInvalidVideoId
		}
	}

	res, err := s.ytClient.Videos.List([]string{"contentDetails"}).
		Id(videoIds...).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("Videos.List: %w", err)
	}

	videos := make(map[string]VideoInfo, len(res.Items))
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
		videos[item.Id] = VideoInfo{
			Id:            item.Id,
			LengthSeconds: lengthSeconds,
		}
	}

	return videos, nil
}

func (s *youTubeAPIServiceImpl) FetchChannelInfo(ctx context.Context, channelId string) (*ChannelInfo, error) {
	q := s.ytClient.Channels.List([]string{
		"snippet",
		"statistics",
	}).Context(ctx)
	if strings.HasPrefix(channelId, "@") && len([]rune(channelId)) > 3 {
		q = q.ForHandle(channelId)
	} else if strings.HasPrefix(channelId, "UC") && len(channelId) == 24 { // NOTE: UCで始まるIDはASCIIのため、runeは使わない
		q = q.Id(channelId)
	} else {
		return nil, ErrInvalidChannelId
	}
	res, err := q.Do()
	if err != nil {
		return nil, fmt.Errorf("Channels.List: %w", err)
	}
	if len(res.Items) == 0 {
		return nil, ErrChannelNotFound
	}
	found := res.Items[0]
	if found.Snippet == nil {
		return nil, ErrInvalidChannelSnippetResponse
	}

	createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}

	iconUrl := ""
	if found.Snippet.Thumbnails != nil {
		if found.Snippet.Thumbnails.Medium != nil {
			iconUrl = found.Snippet.Thumbnails.Medium.Url
		} else if found.Snippet.Thumbnails.Default != nil {
			iconUrl = found.Snippet.Thumbnails.Default.Url
		}
	}
	if iconUrl == "" {
		slog.Warn("icon url not foud when fetch channel info", "channel id", channelId)
	}

	subscribersCount := 0
	if found.Statistics != nil {
		subscribersCount = int(found.Statistics.SubscriberCount)
	}

	return &ChannelInfo{
		Id:               found.Id,
		DisplayName:      found.Snippet.Title,
		CustomId:         found.Snippet.CustomUrl,
		Description:      found.Snippet.Description,
		IconUrl:          iconUrl,
		SubscribersCount: subscribersCount,
		CreatedAt:        createdAt,
	}, nil
}

func (s *youTubeAPIServiceImpl) FetchRSSFeed(ctx context.Context, channelId string) ([]RSSFeedVideo, error) {
	if !strings.HasPrefix(channelId, "UC") || len(channelId) != 24 {
		return nil, ErrInvalidChannelId
	}

	feed, err := s.feedParser.ParseURLWithContext(
		fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelId),
		ctx,
	)
	if err != nil {
		return nil, fmt.Errorf("ParseURLWithContext(channel=%s): %w", channelId, err)
	}

	videos := make([]RSSFeedVideo, 0, len(feed.Items))
	for _, item := range feed.Items {
		video := RSSFeedVideo{
			Title: item.Title,
		}

		if item.PublishedParsed != nil {
			video.CreatedAt = *item.PublishedParsed
		}

		// yt:videoId
		if ytExt, ok := item.Extensions["yt"]; ok {
			if videoIdExt, ok := ytExt["videoId"]; ok && len(videoIdExt) > 0 {
				video.VideoId = videoIdExt[0].Value
			}
		}

		// media:group 内の thumbnail, description, statistics
		if mediaExt, ok := item.Extensions["media"]; ok {
			if groups, ok := mediaExt["group"]; ok && len(groups) > 0 {
				group := groups[0]

				if thumbnails, ok := group.Children["thumbnail"]; ok && len(thumbnails) > 0 {
					video.ThumbnailUrl = thumbnails[0].Attrs["url"]
				}

				if descriptions, ok := group.Children["description"]; ok && len(descriptions) > 0 {
					video.Description = descriptions[0].Value
				}
			}
		}

		videos = append(videos, video)
	}

	return videos, nil
}

func NewYouTubeAPIServiceImpl(ctx context.Context, apiKey string) (YouTubeAPIService, error) {
	ytClient, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("NewService: %w", err)
	}

	feedParser := gofeed.NewParser()
	feedParser.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"

	return &youTubeAPIServiceImpl{
		ytClient:   ytClient,
		feedParser: feedParser,
	}, nil
}
