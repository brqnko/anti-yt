package youtube_d

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
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
	FetchVideoDetail(ctx context.Context, videoIDs []VideoID) (map[VideoID]Video, error)
	FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error)
	SearchVideoIDs(ctx context.Context, query string, pageToken string, opts SearchOptions) (_ []VideoID, _ string, err error)
	OAuthAuthCodeURL(state string) string
	OAuthExchange(ctx context.Context, code string) (accessToken string, err error)
	FetchAllSubscriptions(ctx context.Context, accessToken string) (_ []Channel, err error)
	FetchWatchHistory(ctx context.Context, accessToken string, pageToken string) (_ []WatchHistory, _ string, err error)
	FetchPlaylistVideoIDsWithOAuth(ctx context.Context, accessToken string, playlistID string, pageToken string) (_ []VideoID, _ string, err error)
}

type SearchOptions struct {
	Language          *string
	Order             *string
	PublishedBefore   *time.Time
	PublishedAfter    *time.Time
	RegionCode        *string
	RelevanceLanguage *string
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	ytClient    *youtube.Service
	oauthConfig *oauth2.Config

	lastCheckedDay  time.Time
	consumedQuota   int
	dailyQuotaLimit int
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

	if err := s.tryConsumeQuota(1); err != nil {
		return nil, err
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.ContentDetails is nil(fetchVideoDetail)")
			continue
		}

		lengthSeconds := 0
		matches := iso8601DurationRe.FindStringSubmatch(item.ContentDetails.Duration)
		if matches == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "duration matches is nil(fetchVideoDetail)")
			continue
		}
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		seconds, _ := strconv.Atoi(matches[3])
		lengthSeconds = hours*3600 + minutes*60 + seconds

		if item.Snippet == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Snippet is nil(fetchVideoDetail)")
			continue
		}

		if item.Snippet.Thumbnails == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Snippet.Thumbnails is nil(fetchVideoDetail)")
			continue
		}
		var thumbnail string
		if item.Snippet.Thumbnails.Medium != nil {
			thumbnail = item.Snippet.Thumbnails.Medium.Url
		} else if item.Snippet.Thumbnails.Default != nil {
			thumbnail = item.Snippet.Thumbnails.Default.Url
		} else {
			util.LoggerFromContext(ctx).InfoContext(ctx, "no valid thumbnail in item.Snippet.Thumbnails(fetchVideoDetail)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse createdAt(fetchVideoDetail)")
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to newVideo(fetchVideoDetail)", slog.Any("error", err))
			continue
		}
		videos[video.ID] = video
	}

	return videos, nil
}

func (s *serviceImpl) FetchChannelDetailByIDOrHandle(ctx context.Context, channelID string) (_ Channel, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchChannelDetailByIDOrHandle")

	if err := s.tryConsumeQuota(1); err != nil {
		return Channel{}, err
	}

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

	if err := s.tryConsumeQuota(1); err != nil {
		return nil, err
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.Snippet == nil(fetchChannelDetail)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse createdAt", slog.Any("error", err))
			continue
		}

		if found.Snippet.Thumbnails == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.Snippet.Thumbnails == nil(fetchVideoDetail)")
			continue
		}
		iconURL := ""
		if found.Snippet.Thumbnails.Medium != nil {
			iconURL = found.Snippet.Thumbnails.Medium.Url
		} else if found.Snippet.Thumbnails.Default != nil {
			iconURL = found.Snippet.Thumbnails.Default.Url
		} else {
			util.LoggerFromContext(ctx).InfoContext(ctx, "no valid iconURL found(fetchVideoDetail)")
			continue
		}

		if found.Statistics == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.Statistics == nil(fetchVideoDetail)")
			continue
		}

		if found.ContentDetails == nil || found.ContentDetails.RelatedPlaylists == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.ContentDetails or found.ContentDetails.RelatedPlailits is nil(fetchVideoDetail)")
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to NewChannel(fetchVideoDetail)", slog.Any("error", err))
			continue
		}

		result[channel.ID] = channel
	}

	return result, nil
}

func (s *serviceImpl) FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "youTubeAPIService.FetchPlaylistVideoIDs")

	if err := s.tryConsumeQuota(1); err != nil {
		return nil, "", err
	}

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
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.ContentDetails is nil or VideoId is empty(fetchPlaylistVideoIDs)")
			continue
		}
		videoID, err := NewVideoID(item.ContentDetails.VideoId)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to NewVideoID(fetchPlaylistVideoIDs)", slog.Any("error", err))
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}

func (s *serviceImpl) SearchVideoIDs(ctx context.Context, query string, pageToken string, opts SearchOptions) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "youTubeAPIService.SearchVideoIDs")

	if err := s.tryConsumeQuota(100); err != nil {
		return nil, "", err
	}

	call := s.ytClient.Search.List([]string{"id"}).
		Q(query).
		Type("video").
		VideoDuration("medium").
		MaxResults(50).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if opts.Language != nil && *opts.Language != "" {
		call = call.RelevanceLanguage(*opts.Language)
	}
	if opts.Order != nil && *opts.Order != "" {
		call = call.Order(*opts.Order)
	}
	if opts.PublishedBefore != nil {
		call = call.PublishedBefore(opts.PublishedBefore.Format(time.RFC3339))
	}
	if opts.PublishedAfter != nil {
		call = call.PublishedAfter(opts.PublishedAfter.Format(time.RFC3339))
	}
	if opts.RegionCode != nil && *opts.RegionCode != "" {
		call = call.RegionCode(*opts.RegionCode)
	}
	if opts.RelevanceLanguage != nil && *opts.RelevanceLanguage != "" {
		call = call.RelevanceLanguage(*opts.RelevanceLanguage)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	videoIDs := make([]VideoID, 0, len(res.Items))
	for _, item := range res.Items {
		if item.Id == nil || item.Id.VideoId == "" {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Id is nil or VideoId is empty(searchVideoIDs)")
			continue
		}
		videoID, err := NewVideoID(item.Id.VideoId)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to NewVideoID(searchVideoIDs)", slog.Any("error", err))
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}

var (
	errDailyQuotaExceeded = errors.New("daily quota exceeded")
)

func (s *serviceImpl) tryConsumeQuota(quota int) error {
	if s.lastCheckedDay.Before(quotaDate()) {
		s.consumedQuota = 0
		s.lastCheckedDay = quotaDate()
	}

	if s.consumedQuota+quota > s.dailyQuotaLimit {
		return errDailyQuotaExceeded
	}

	s.consumedQuota += quota
	return nil
}

func NewService(ctx context.Context, apiKey, oauthClientID, oauthClientSecret, oauthRedirectURL string) (_ Service, err error) {
	defer util.Wrap(&err, "NewYouTubeAPIServiceImpl")

	ytClient, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &serviceImpl{
		ytClient: ytClient,
		oauthConfig: &oauth2.Config{
			ClientID:     oauthClientID,
			ClientSecret: oauthClientSecret,
			RedirectURL:  oauthRedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/youtube.readonly"},
			Endpoint:     googleOAuth2.Endpoint,
		},
		lastCheckedDay:  quotaDate().Add(24 * time.Hour),
		consumedQuota:   0,
		dailyQuotaLimit: 10000,
	}, nil
}

func (s *serviceImpl) OAuthAuthCodeURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *serviceImpl) OAuthExchange(ctx context.Context, code string) (_ string, err error) {
	defer util.Wrap(&err, "serviceImpl.OAuthExchange")

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (s *serviceImpl) FetchAllSubscriptions(ctx context.Context, accessToken string) (_ []Channel, err error) {
	defer util.Wrap(&err, "serviceImpl.FetchAllSubscriptions")

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
	ytClient, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	var channels []Channel
	pageToken := ""
	for {
		call := ytClient.Subscriptions.List([]string{"snippet"}).Mine(true).MaxResults(50)
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


func (s *serviceImpl) FetchPlaylistVideoIDsWithOAuth(ctx context.Context, accessToken string, playlistID string, pageToken string) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "serviceImpl.FetchPlaylistVideoIDsWithOAuth")

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
	ytClient, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, "", err
	}

	call := ytClient.PlaylistItems.List([]string{"contentDetails"}).
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.ContentDetails is nil or VideoId is empty(fetchPlaylistVideoIDsWithOAuth)")
			continue
		}
		videoID, err := NewVideoID(item.ContentDetails.VideoId)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to NewVideoID(fetchPlaylistVideoIDsWithOAuth)", slog.Any("error", err))
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}

func (s *serviceImpl) FetchWatchHistory(ctx context.Context, accessToken string, pageToken string) (_ []WatchHistory, _ string, err error) {
	defer util.Wrap(&err, "serviceImpl.FetchWatchHistory")

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
	ytClient, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, "", err
	}

	call := ytClient.PlaylistItems.List([]string{"snippet"}).
		PlaylistId("HL").
		MaxResults(50).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	histories := make([]WatchHistory, 0, len(res.Items))
	for _, item := range res.Items {
		if item.Snippet == nil || item.Snippet.ResourceId == nil || item.Snippet.ResourceId.VideoId == "" {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Snippet or ResourceId is nil or VideoId is empty(fetchWatchHistory)")
			continue
		}
		videoID, err := NewVideoID(item.Snippet.ResourceId.VideoId)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to NewVideoID(fetchWatchHistory)", slog.Any("error", err))
			continue
		}
		watchedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse PublishedAt(fetchWatchHistory)", slog.Any("error", err))
			continue
		}
		histories = append(histories, WatchHistory{
			VideoID:   videoID,
			WatchedAt: watchedAt,
		})
	}

	return histories, res.NextPageToken, nil
}

func quotaDate() time.Time {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	y, m, d := time.Now().In(loc).Date()

	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}
