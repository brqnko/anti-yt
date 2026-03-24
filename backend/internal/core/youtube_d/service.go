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

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/mmcdole/gofeed"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var iso8601DurationRe = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

var (
	ErrInvalidChannelID = core.NewDomainError("youtube.invalid_channel_id", "invalid channel id")

	ErrVideoIDsTooMuch = core.NewDomainError("youtube.video_ids_too_much", "video ids are too much")
)

type Service interface {
	FetchChannelDetail(ctx context.Context, channelIDs []ChannelID) (map[ChannelID]Channel, error)
	FetchChannelDetailByIDOrHandle(ctx context.Context, channelID string) (Channel, error)
	FetchRSSFeed(ctx context.Context, channelID ChannelID) ([]VideoID, error)
	FetchVideoDetail(ctx context.Context, videoIDs []VideoID) (map[VideoID]Video, error)
	FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error)
	SearchVideoIDs(ctx context.Context, query string, pageToken string) (_ []VideoID, _ string, err error)
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	ytClient   *youtube.Service
	feedParser *gofeed.Parser
}

// NOTE: 削除された動画などはAPIから返ってこず、また順番も保証されないのでmapで返す
func (s *serviceImpl) FetchVideoDetail(ctx context.Context, videoIDs []VideoID) (_ map[VideoID]Video, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchVideoDetail")

	if len(videoIDs) == 0 {
		return map[VideoID]Video{}, nil
	}
	if len(videoIDs) > 50 {
		return nil, ErrVideoIDsTooMuch
	}

	ids := make([]string, len(videoIDs))
	for i, id := range videoIDs {
		ids[i] = (string)(id)
	}
	res, err := s.ytClient.Videos.List([]string{"snippet", "contentDetails"}).
		Id(ids...).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	videos := make(map[VideoID]Video)
	for _, item := range res.Items {
		if item.ContentDetails == nil {
			slog.Info("item.ContentDetails is nil(fetchVideoDetail)")
			continue
		}

		lengthSeconds := 0
		matches := iso8601DurationRe.FindStringSubmatch(item.ContentDetails.Duration)
		if matches == nil {
			slog.Info("duration matches is nil(fetchVideoDetail)")
			continue
		}
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		seconds, _ := strconv.Atoi(matches[3])
		lengthSeconds = hours*3600 + minutes*60 + seconds

		if item.Snippet == nil {
			slog.Info("item.Snippet is nil(fetchVideoDetail)")
			continue
		}

		if item.Snippet.Thumbnails == nil {
			slog.Info("item.Snippet.Thumbnails is nil(fetchVideoDetail)")
			continue
		}
		var thumbnail string
		if item.Snippet.Thumbnails.Medium != nil {
			thumbnail = item.Snippet.Thumbnails.Medium.Url
		} else if item.Snippet.Thumbnails.Default != nil {
			thumbnail = item.Snippet.Thumbnails.Default.Url
		} else {
			slog.Info("no valid thumbnail in item.Snippet.Thumbnails(fetchVideoDetail)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			slog.Info("failed to parse createdAt(fetchVideoDetail)")
			continue
		}

		video, err := NewVideo(
			item.Id,
			item.Snippet.ChannelId,
			item.Snippet.Title,
			item.Snippet.Description,
			thumbnail,
			lengthSeconds,
			createdAt,
		)
		if err != nil {
			slog.Info("failed to newVideo(fetchVideoDetail)", "error", err)
			continue
		}
		videos[video.ID] = video
	}

	return videos, nil
}

func (s *serviceImpl) FetchChannelDetailByIDOrHandle(ctx context.Context, channelID string) (_ Channel, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchChannelDetailByIDOrHandle")

	q := s.ytClient.Channels.List([]string{"snippet", "statistics", "contentDetails"})
	if strings.HasPrefix(channelID, "@") && len(([]rune)(channelID)) > 3 {
		q = q.ForHandle(channelID)
	} else if strings.HasPrefix(channelID, "UC") && len(channelID) == 24 {
		q = q.Id(channelID)
	} else {
		return Channel{}, ErrInvalidChannelID
	}

	res, err := q.Context(ctx).Do()
	if err != nil {
		return Channel{}, err
	}
	if len(res.Items) == 0 {
		return Channel{}, errors.New("len(res.Items) == 0(fetchChannelDetail)")
	}
	found := res.Items[0]

	if found.Snippet == nil {
		return Channel{}, errors.New("found.Snippet == nil(fetchChannelDetail)")
	}

	createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
	if err != nil {
		return Channel{}, errors.New("failed to parse createdAt(fetchChannelDetail)")
	}

	if found.Snippet.Thumbnails == nil {
		return Channel{}, errors.New("found.Snippet.Thumbnails == nil(fetchChannelDetail)")
	}
	iconURL := ""
	if found.Snippet.Thumbnails.Medium != nil {
		iconURL = found.Snippet.Thumbnails.Medium.Url
	} else if found.Snippet.Thumbnails.Default != nil {
		iconURL = found.Snippet.Thumbnails.Default.Url
	} else {
		return Channel{}, errors.New("no valid iconURL found(fetchChannelDetail)")
	}

	if found.Statistics == nil {
		return Channel{}, errors.New("found.Statistics == nil(fetchChannelDetail)")
	}
	if found.ContentDetails == nil || found.ContentDetails.RelatedPlaylists == nil {
		return Channel{}, errors.New("found.ContentDetails or found.ContentDetails.ReleatedPlaylists == nil(fetchVideoDetail)")
	}

	channel, err := NewChannel(
		found.Id,
		found.Snippet.Title,
		found.Snippet.CustomUrl,
		found.Snippet.Description,
		iconURL,
		found.Statistics.SubscriberCount,
		found.ContentDetails.RelatedPlaylists.Uploads,
		createdAt,
	)
	if err != nil {
		return Channel{}, err
	}

	return channel, nil
}

func (s *serviceImpl) FetchChannelDetail(ctx context.Context, channelIDs []ChannelID) (_ map[ChannelID]Channel, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchChannelDetail")

	if len(channelIDs) == 0 {
		return map[ChannelID]Channel{}, nil
	}

	result := make(map[ChannelID]Channel)
	ids := make([]string, len(channelIDs))
	for i, channelID := range channelIDs {
		ids[i] = (string)(channelID)
	}
	res, err := s.ytClient.Channels.List([]string{"snippet", "statistics", "contentDetails"}).Id(ids...).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	for _, found := range res.Items {
		if found.Snippet == nil {
			slog.Info("found.Snippet == nil(fetchChannelDetail)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
		if err != nil {
			slog.Info("failed to parse createdAt", "error", err)
			continue
		}

		if found.Snippet.Thumbnails == nil {
			slog.Info("found.Snippet.Thumbnails == nil(fetchVideoDetail)")
			continue
		}
		iconURL := ""
		if found.Snippet.Thumbnails.Medium != nil {
			iconURL = found.Snippet.Thumbnails.Medium.Url
		} else if found.Snippet.Thumbnails.Default != nil {
			iconURL = found.Snippet.Thumbnails.Default.Url
		} else {
			slog.Info("no valid iconURL found(fetchVideoDetail)")
			continue
		}

		if found.Statistics == nil {
			slog.Info("found.Statistics == nil(fetchVideoDetail)")
			continue
		}

		if found.ContentDetails == nil || found.ContentDetails.RelatedPlaylists == nil {
			slog.Info("found.ContentDetails or found.ContentDetails.RelatedPlailits is nil(fetchVideoDetail)")
			continue
		}

		channel, err := NewChannel(
			found.Id,
			found.Snippet.Title,
			found.Snippet.CustomUrl,
			found.Snippet.Description,
			iconURL,
			found.Statistics.SubscriberCount,
			found.ContentDetails.RelatedPlaylists.Uploads,
			createdAt,
		)
		if err != nil {
			slog.Info("failed to NewChannel(fetchVideoDetail)", "error", err)
			continue
		}

		result[channel.ID] = channel
	}

	return result, nil
}

func (s *serviceImpl) FetchRSSFeed(ctx context.Context, channelID ChannelID) (_ []VideoID, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchRSSFeed")

	if !strings.HasPrefix((string)(channelID), "UC") || len(channelID) != 24 {
		return nil, ErrInvalidChannelID
	}

	feed, err := s.feedParser.ParseURLWithContext(
		fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID),
		ctx,
	)
	if err != nil {
		return nil, err
	}

	videos := make([]VideoID, 0, len(feed.Items))
	for _, item := range feed.Items {
		// yt:videoId
		ytExt, ok := item.Extensions["yt"]
		if !ok {
			slog.Info("yt extension not found(fetchRSSFeed)")
			continue
		}
		videoIDExt, ok := ytExt["videoId"]
		if !ok || len(videoIDExt) == 0 {
			slog.Info("yt:videoId not found(fetchRSSFeed)")
			continue
		}

		videoID, err := NewVideoID(videoIDExt[0].Value)
		if err != nil {
			slog.Info("failed to NewRSSFeedVideo(fetchRSSFeed)", "error", err)
			continue
		}
		videos = append(videos, videoID)
	}

	return videos, nil
}

func (s *serviceImpl) FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchPlaylistVideoIDs")

	call := s.ytClient.PlaylistItems.List([]string{"contentDetails"}).
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
			slog.Info("item.ContentDetails is nil or VideoId is empty(fetchPlaylistVideoIDs)")
			continue
		}
		videoID, err := NewVideoID(item.ContentDetails.VideoId)
		if err != nil {
			slog.Info("failed to NewVideoID(fetchPlaylistVideoIDs)", "error", err)
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}

func (s *serviceImpl) SearchVideoIDs(ctx context.Context, query string, pageToken string) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "youTubeAPIService.SearchVideoIDs")

	call := s.ytClient.Search.List([]string{"id"}).
		Q(query).
		Type("video").
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
		if item.Id == nil || item.Id.VideoId == "" {
			slog.Info("item.Id is nil or VideoId is empty(searchVideoIDs)")
			continue
		}
		videoID, err := NewVideoID(item.Id.VideoId)
		if err != nil {
			slog.Info("failed to NewVideoID(searchVideoIDs)", "error", err)
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}

func NewService(ctx context.Context, apiKey string) (_ Service, err error) {
	defer util.Wrap(&err, "NewYouTubeAPIServiceImpl")

	ytClient, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	feedParser := gofeed.NewParser()
	feedParser.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"

	return &serviceImpl{
		ytClient:   ytClient,
		feedParser: feedParser,
	}, nil
}
