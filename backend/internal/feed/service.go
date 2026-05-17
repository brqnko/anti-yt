package feed

//go:generate moq -out mock_youtube_client_test.go -pkg feed_test ../core/youtube_d Client

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db            *pgxpool.Pool
	youtubeClient youtube_d.Client
	feedRepo      database_d.FeedRepository
	feedQS        FeedQueryService

	rssFetchDuration time.Duration
}

func NewService(db *pgxpool.Pool, youtubeClient youtube_d.Client, feedRepo database_d.FeedRepository, rssFetchDuration time.Duration) *Service {
	return new(Service{
		db:               db,
		youtubeClient:    youtubeClient,
		feedRepo:         feedRepo,
		feedQS:           NewFeedQueryService(db),
		rssFetchDuration: rssFetchDuration,
	})
}

func (s *Service) GetFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetVideoFeedView, _ bool, err error) {
	defer util.Wrap(&err, "feed.(*Service).GetFeed")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).WarnContext(ctx, "failed to rollback transaction", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// NOTE: YouTube RSS Feedはレートリミットが厳しいのでlimit=1
	// rss_fetched_atで昇順ソートしているので、ちょっとずつ更新されていくはず
	channels, err := channel.NewChannelRepository(q).FindToFetchRSSForUpdate(ctx, userID, s.rssFetchDuration, 1)
	if err != nil {
		return nil, false, err
	}

	for _, ch := range channels {
		// PlaylistAPIから動画ID一覧を取得する
		videoIDs, _, err := s.youtubeClient.FetchPlaylistVideoIDs(ctx, string(ch.Channel.UploadsPlaylistID), "")
		if err != nil {
			return nil, false, err
		}

		// 動画ID一覧から動画の詳細情報を取得する
		// TODO: 現在はchannels分YouTube Data APIにリクエストを投げているが、一括で取得したい
		videoDetailMap, err := s.youtubeClient.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			return nil, false, err
		}

		fetchedAt := time.Now().UTC()
		for _, vd := range videoDetailMap {
			v, err := video.NewVideo(ch.ID, fetchedAt, vd)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(get feed)", slog.Any("error", err))
				continue
			}

			if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(get feed)", slog.Any("error", err))
				continue
			}
			if err := s.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
				util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(get feed)", slog.Any("error", err))
			}
		}

		ch.MarkAsRSSFetched()
		if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	videoIDs, err := s.feedRepo.Get(ctx, userID, cursor, int64(limit)+1)
	if err != nil {
		return nil, false, err
	}

	videos, err := s.feedQS.HydrateVideoFeed(ctx, userID, videoIDs)
	if err != nil {
		return nil, false, err
	}

	if len(videos) > int(limit) {
		return videos[:limit], true, nil
	}
	return videos, false, nil
}

func (s *Service) GetLatestVideos(ctx context.Context, cursor *uuid.UUID, limit int32) (_ []GetVideoFeedView, _ bool, err error) {
	defer util.Wrap(&err, "feed.(*Service).GetLatestVideos")

	videos, err := s.feedQS.ListLatestVideos(ctx, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(videos) > int(limit) {
		return videos[:limit], true, nil
	}
	return videos, false, nil
}

type SearchItemView struct {
	Type string // "video" | "channel"

	// 動画・チャンネル共通
	ChannelID                  uuid.UUID
	ExternalChannelID          string
	ExternalChannelDisplayName string
	ExternalChannelIconUrl     string

	// 動画のみ
	VideoID                    uuid.UUID
	ExternalVideoID            string
	ExternalVideoTitle         string
	ExternalVideoDescription   string
	ExternalVideoThumbnailUrl  string
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int

	// チャンネルのみ
	ChannelCustomID         string
	ChannelSubscribersCount int
}

func (s *Service) Search(ctx context.Context, query string, limit int, cursor *string, opts youtube_d.SearchOptions) (_ []SearchItemView, _ bool, _ *string, err error) {
	defer util.Wrap(&err, "feed.(*Service).Search")

	pageToken := ""
	if cursor != nil {
		pageToken = *cursor
	}

	searchItems, nextPageToken, err := s.youtubeClient.SearchIDs(ctx, query, pageToken, opts)
	if err != nil {
		return nil, false, nil, err
	}

	if len(searchItems) == 0 {
		return []SearchItemView{}, false, nil, nil
	}

	// 動画IDとチャンネルIDを分類
	var videoIDs []youtube_d.VideoID
	var channelSearchIDs []youtube_d.ChannelID
	for _, item := range searchItems {
		switch item.Type {
		case youtube_d.SearchItemTypeVideo:
			videoIDs = append(videoIDs, item.VideoID)
		case youtube_d.SearchItemTypeChannel:
			channelSearchIDs = append(channelSearchIDs, item.ChannelID)
		}
	}

	fetchedAt := time.Now().UTC()
	q := sqlc.New(s.db)

	// 動画詳細・チャンネル詳細を取得
	var videoDetails map[youtube_d.VideoID]youtube_d.Video
	var videoChannelDetails map[youtube_d.ChannelID]youtube_d.Channel
	if len(videoIDs) > 0 {
		videoDetails, err = s.youtubeClient.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			return nil, false, nil, err
		}
		vcIDs := make([]youtube_d.ChannelID, 0, len(videoDetails))
		for _, vd := range videoDetails {
			vcIDs = append(vcIDs, vd.ChannelID)
		}
		videoChannelDetails, err = s.youtubeClient.FetchChannelDetail(ctx, vcIDs)
		if err != nil {
			return nil, false, nil, err
		}
	}

	var channelSearchDetails map[youtube_d.ChannelID]youtube_d.Channel
	if len(channelSearchIDs) > 0 {
		channelSearchDetails, err = s.youtubeClient.FetchChannelDetail(ctx, channelSearchIDs)
		if err != nil {
			return nil, false, nil, err
		}
	}

	// チャンネルをまとめて保存（動画用 + チャンネル検索用）
	savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
	allChannelDetails := make(map[youtube_d.ChannelID]youtube_d.Channel)
	for k, v := range videoChannelDetails {
		allChannelDetails[k] = v
	}
	for k, v := range channelSearchDetails {
		allChannelDetails[k] = v
	}
	for _, cd := range allChannelDetails {
		ch, err := channel.NewChannel(fetchedAt, fetchedAt.AddDate(-1, 0, 0), cd)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new channel(search)", slog.Any("error", err))
			continue
		}
		if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save channel(search)", slog.Any("error", err))
			continue
		}
		savedChannels[cd.ID] = ch.ID
	}

	// 検索順序を維持しながら結果を構築
	items := make([]SearchItemView, 0, len(searchItems))
	for _, si := range searchItems {
		switch si.Type {
		case youtube_d.SearchItemTypeVideo:
			vd, ok := videoDetails[si.VideoID]
			if !ok {
				continue
			}
			channelUUID, ok := savedChannels[vd.ChannelID]
			if !ok {
				continue
			}
			cd, ok := allChannelDetails[vd.ChannelID]
			if !ok {
				continue
			}
			v, err := video.NewVideo(channelUUID, fetchedAt, vd)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(search)", slog.Any("error", err))
				continue
			}
			if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(search)", slog.Any("error", err))
				continue
			}
			if err := s.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
				util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(search)", slog.Any("error", err))
			}
			items = append(items, SearchItemView{
				Type:                       "video",
				VideoID:                    v.ID,
				ChannelID:                  channelUUID,
				ExternalVideoID:            string(vd.ID),
				ExternalChannelID:          string(vd.ChannelID),
				ExternalVideoTitle:         vd.Title,
				ExternalVideoDescription:   vd.Description,
				ExternalVideoThumbnailUrl:  vd.ThumbnailURL,
				ExternalChannelDisplayName: cd.DisplayName,
				ExternalChannelIconUrl:     cd.IconURL,
				ExternalVideoCreatedAt:     vd.CreatedAt,
				ExternalVideoLengthSeconds: vd.LengthSeconds,
			})

		case youtube_d.SearchItemTypeChannel:
			cd, ok := allChannelDetails[si.ChannelID]
			if !ok {
				continue
			}
			channelUUID, ok := savedChannels[si.ChannelID]
			if !ok {
				continue
			}
			items = append(items, SearchItemView{
				Type:                       "channel",
				ChannelID:                  channelUUID,
				ExternalChannelID:          string(cd.ID),
				ExternalChannelDisplayName: cd.DisplayName,
				ExternalChannelIconUrl:     cd.IconURL,
				ChannelCustomID:            cd.CustomID,
				ChannelSubscribersCount:    int(cd.SubscribersCount),
			})
		}
	}

	hasNext := nextPageToken != ""
	var nextCursor *string
	if hasNext {
		nextCursor = &nextPageToken
	}

	if len(items) > limit {
		items = items[:limit]
	}

	return items, hasNext, nextCursor, nil
}
