package youtube_d

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
	"google.golang.org/api/youtube/v3"
)

var iso8601DurationRe = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

var (
	ErrInvalidChannelID = core.NewDomainError("youtube.invalid_channel_id", "invalid channel id", core.StatusBadRequest)

	ErrVideoIDsTooMuch = core.NewDomainError("youtube.video_ids_too_much", "video ids are too much", core.StatusBadRequest)

	ErrQuotaExceeded = errors.New("daily quota exceeded")
)

type Client interface {
	FetchChannelDetail(ctx context.Context, channelIDs []ChannelID) (map[ChannelID]Channel, error)
	FetchChannelDetailByIDOrHandle(ctx context.Context, channelID string) (Channel, error)
	FetchVideoDetail(ctx context.Context, videoIDs []VideoID) (map[VideoID]Video, error)
	FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error)
	FetchChannelPlaylists(ctx context.Context, channelID ChannelID, pageToken string) (_ []Playlist, _ string, err error)
	SearchIDs(ctx context.Context, query string, pageToken string, opts SearchOptions) (_ []SearchItem, _ string, err error)
	OAuthAuthCodeURL(state string) string
	OAuthExchange(ctx context.Context, code string) (*OAuthClient, error)
}

type SearchOptions struct {
	Language          *string
	Order             *string
	PublishedBefore   *time.Time
	PublishedAfter    *time.Time
	RegionCode        *string
	RelevanceLanguage *string
}

var _ Client = (*clientImpl)(nil)

type clientImpl struct {
	ytClient    *youtube.Service
	oauthConfig *oauth2.Config

	mu             sync.RWMutex
	lastCheckedDay time.Time
	quotaExceeded  bool
}

// NOTE: 削除された動画などはAPIから返ってこず、また順番も保証されないのでmapで返す
func (s *clientImpl) FetchVideoDetail(ctx context.Context, videoIDs []VideoID) (_ map[VideoID]Video, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).FetchVideoDetail")
	defer s.markIfQuotaExceeded(&err)

	if len(videoIDs) == 0 {
		return map[VideoID]Video{}, nil
	}
	if len(videoIDs) > 50 {
		return nil, ErrVideoIDsTooMuch
	}

	if err := s.checkQuota(); err != nil {
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.ContentDetails is nil(fetch video detail)")
			continue
		}

		lengthSeconds := 0
		matches := iso8601DurationRe.FindStringSubmatch(item.ContentDetails.Duration)
		if matches != nil {
			hours, _ := strconv.Atoi(matches[1])
			minutes, _ := strconv.Atoi(matches[2])
			seconds, _ := strconv.Atoi(matches[3])
			lengthSeconds = hours*3600 + minutes*60 + seconds
		}

		if item.Snippet == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Snippet is nil(fetch video detail)")
			continue
		}

		if item.Snippet.Thumbnails == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Snippet.Thumbnails is nil(fetch video detail)")
			continue
		}
		var thumbnail string
		if item.Snippet.Thumbnails.Medium != nil {
			thumbnail = item.Snippet.Thumbnails.Medium.Url
		} else if item.Snippet.Thumbnails.Default != nil {
			thumbnail = item.Snippet.Thumbnails.Default.Url
		} else {
			util.LoggerFromContext(ctx).InfoContext(ctx, "no valid thumbnail in item.Snippet.Thumbnails(fetch video detail)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse created at(fetch video detail)")
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(fetch video detail)", slog.Any("error", err))
			continue
		}
		videos[video.ID] = video
	}

	return videos, nil
}

func (s *clientImpl) FetchChannelDetailByIDOrHandle(ctx context.Context, channelID string) (_ Channel, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).FetchChannelDetailByIDOrHandle")
	defer s.markIfQuotaExceeded(&err)

	if err := s.checkQuota(); err != nil {
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
		return Channel{}, errors.New("res.Items is empty")
	}
	found := res.Items[0]

	if found.Snippet == nil {
		return Channel{}, errors.New("found.Snippet is nil")
	}

	createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
	if err != nil {
		return Channel{}, err
	}

	if found.Snippet.Thumbnails == nil {
		return Channel{}, errors.New("found.Snippet.Thumbnails is nil")
	}
	var iconURL string
	if found.Snippet.Thumbnails.Medium != nil {
		iconURL = found.Snippet.Thumbnails.Medium.Url
	} else if found.Snippet.Thumbnails.Default != nil {
		iconURL = found.Snippet.Thumbnails.Default.Url
	} else {
		return Channel{}, errors.New("no valid iconURL found")
	}

	if found.Statistics == nil {
		return Channel{}, errors.New("found.Statistics is nil")
	}
	if found.ContentDetails == nil || found.ContentDetails.RelatedPlaylists == nil {
		return Channel{}, errors.New("found.ContentDetails or found.ContentDetails.RelatedPlaylists is nil")
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

func (s *clientImpl) FetchChannelDetail(ctx context.Context, channelIDs []ChannelID) (_ map[ChannelID]Channel, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).FetchChannelDetail")
	defer s.markIfQuotaExceeded(&err)

	if len(channelIDs) == 0 {
		return map[ChannelID]Channel{}, nil
	}

	if err := s.checkQuota(); err != nil {
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.Snippet is nil(fetch channel detail)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, found.Snippet.PublishedAt)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse created at(fetch channel detail)", slog.Any("error", err))
			continue
		}

		if found.Snippet.Thumbnails == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.Snippet.Thumbnails is nil(fetch channel detail)")
			continue
		}
		var iconURL string
		if found.Snippet.Thumbnails.Medium != nil {
			iconURL = found.Snippet.Thumbnails.Medium.Url
		} else if found.Snippet.Thumbnails.Default != nil {
			iconURL = found.Snippet.Thumbnails.Default.Url
		} else {
			util.LoggerFromContext(ctx).InfoContext(ctx, "no valid icon url found(fetch channel detail)")
			continue
		}

		if found.Statistics == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.Statistics is nil(fetch channel detail)")
			continue
		}

		if found.ContentDetails == nil || found.ContentDetails.RelatedPlaylists == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "found.ContentDetails or RelatedPlaylists is nil(fetch channel detail)")
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new channel(fetch channel detail)", slog.Any("error", err))
			continue
		}

		result[channel.ID] = channel
	}

	return result, nil
}

func (s *clientImpl) FetchPlaylistVideoIDs(ctx context.Context, playlistID string, pageToken string) (_ []VideoID, _ string, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).FetchPlaylistVideoIDs")
	defer s.markIfQuotaExceeded(&err)

	if err := s.checkQuota(); err != nil {
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
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.ContentDetails is nil or VideoId is empty(fetch playlist video ids)")
			continue
		}
		videoID, err := NewVideoID(item.ContentDetails.VideoId)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video id(fetch playlist video ids)", slog.Any("error", err))
			continue
		}
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, res.NextPageToken, nil
}

func (s *clientImpl) FetchChannelPlaylists(ctx context.Context, channelID ChannelID, pageToken string) (_ []Playlist, _ string, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).FetchChannelPlaylists")
	defer s.markIfQuotaExceeded(&err)

	if err := s.checkQuota(); err != nil {
		return nil, "", err
	}

	call := s.ytClient.Playlists.List([]string{"snippet"}).
		ChannelId(string(channelID)).
		MaxResults(50).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	playlists := make([]Playlist, 0, len(res.Items))
	for _, item := range res.Items {
		if item.Snippet == nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "item.Snippet is nil(fetch channel playlists)")
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to parse created at(fetch channel playlists)", slog.Any("error", err))
			continue
		}

		playlists = append(playlists, Playlist{
			ID:          item.Id,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			CreatedAt:   createdAt,
		})
	}

	return playlists, res.NextPageToken, nil
}


func (s *clientImpl) checkQuota() error {
	today := quotaDate()

	s.mu.RLock()
	if !s.lastCheckedDay.Before(today) {
		exceeded := s.quotaExceeded
		s.mu.RUnlock()
		if exceeded {
			return ErrQuotaExceeded
		}
		return nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastCheckedDay.Before(today) {
		s.quotaExceeded = false
		s.lastCheckedDay = today
	}
	if s.quotaExceeded {
		return ErrQuotaExceeded
	}
	return nil
}

func (s *clientImpl) markIfQuotaExceeded(err *error) {
	if err == nil || *err == nil {
		return
	}
	var apiErr *googleapi.Error
	if !errors.As(*err, &apiErr) {
		return
	}
	for _, e := range apiErr.Errors {
		if e.Reason != "quotaExceeded" {
			continue
		}
		s.mu.RLock()
		already := s.quotaExceeded
		s.mu.RUnlock()
		if already {
			return
		}
		s.mu.Lock()
		s.quotaExceeded = true
		s.mu.Unlock()
		return
	}
}

func NewClient(ctx context.Context, apiKey, oauthClientID, oauthClientSecret, oauthRedirectURL string) (_ Client, err error) {
	defer util.Wrap(&err, "youtube_d.NewClient")

	baseTransport, err := htransport.NewTransport(ctx, http.DefaultTransport, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	ytClient, err := youtube.NewService(ctx, option.WithHTTPClient(new(http.Client{Transport: otelhttp.NewTransport(baseTransport)})))
	if err != nil {
		return nil, err
	}

	return &clientImpl{
		ytClient: ytClient,
		oauthConfig: new(oauth2.Config{
			ClientID:     oauthClientID,
			ClientSecret: oauthClientSecret,
			RedirectURL:  oauthRedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/youtube.readonly"},
			Endpoint:     googleOAuth2.Endpoint,
		}),
		lastCheckedDay: quotaDate(),
		quotaExceeded:  false,
	}, nil
}

func (s *clientImpl) OAuthAuthCodeURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *clientImpl) OAuthExchange(ctx context.Context, code string) (_ *OAuthClient, err error) {
	defer util.Wrap(&err, "youtube_d.(*clientImpl).OAuthExchange")

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)})
	yt, err := youtube.NewService(ctx, option.WithHTTPClient(oauth2.NewClient(ctx, s.oauthConfig.TokenSource(ctx, token))))
	if err != nil {
		return nil, err
	}

	return &OAuthClient{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ytClient:     yt,
	}, nil
}

func quotaDate() time.Time {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	y, m, d := time.Now().In(loc).Date()

	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}
