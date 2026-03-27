package feed

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db        *pgxpool.Pool
	ytService youtube_d.Service
	feedQS    FeedQueryService

	rssFetchDuration time.Duration
}

func NewService(db *pgxpool.Pool, ytService youtube_d.Service, rssFetchDuration time.Duration) *Service {
	return &Service{
		db:               db,
		ytService:        ytService,
		feedQS:           NewFeedQueryService(db),
		rssFetchDuration: rssFetchDuration,
	}
}

func (s *Service) GetFeed(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetVideoFeedView, _ bool, err error) {
	defer util.Wrap(&err, "Service.GetFeed")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Warn("failed to rollback transaction", "error", err)
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
		// RSSから動画ID一覧を取得する
		// TODO: チャンネルが削除されてrssの取得に失敗した場合のケースを考慮する
		videoIDs, err := s.ytService.FetchRSSFeed(ctx, ch.Channel.ID)
		if err != nil {
			return nil, false, err
		}

		// 動画ID一覧から動画の詳細情報を取得する
		// TODO: 現在はchannels分YouTube Data APIにリクエストを投げているが、一括で取得したい
		videoDetailMap, err := s.ytService.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			return nil, false, err
		}

		fetchedAt := time.Now().UTC()
		for _, vd := range videoDetailMap {
			v, err := video.NewVideo(ch.ID, fetchedAt, vd)
			if err != nil {
				slog.Info("failed to newVideo", "error", err)
				continue
			}

			if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
				slog.Info("failed to save video", "error", err)
				continue
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

	videos, err := s.feedQS.GetVideoFeed(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(videos) > int(limit) {
		return videos[:limit], true, nil
	}
	return videos, false, nil
}

type SearchVideoView struct {
	VideoID                    uuid.UUID
	ChannelID                  uuid.UUID
	ExternalVideoID            string
	ExternalChannelID          string
	ExternalVideoTitle         string
	ExternalVideoDescription   string
	ExternalVideoThumbnailUrl  string
	ExternalChannelDisplayName string
	ExternalChannelIconUrl     string
	ExternalVideoCreatedAt     time.Time
	ExternalVideoLengthSeconds int
}

func (s *Service) Search(ctx context.Context, query string, limit int, cursor *string, opts youtube_d.SearchOptions) (_ []SearchVideoView, _ bool, _ *string, err error) {
	defer util.Wrap(&err, "Service.Search")

	pageToken := ""
	if cursor != nil {
		pageToken = *cursor
	}

	// YouTube Search APIで動画IDを取得
	videoIDs, nextPageToken, err := s.ytService.SearchVideoIDs(ctx, query, pageToken, opts)
	if err != nil {
		return nil, false, nil, err
	}

	if len(videoIDs) == 0 {
		return []SearchVideoView{}, false, nil, nil
	}

	// 動画の詳細を取得
	videoDetails, err := s.ytService.FetchVideoDetail(ctx, videoIDs)
	if err != nil {
		return nil, false, nil, err
	}

	// チャンネルIDを収集
	channelIDs := make([]youtube_d.ChannelID, 0, len(videoDetails))
	for _, vd := range videoDetails {
		channelIDs = append(channelIDs, vd.ChannelID)
	}

	// チャンネルの詳細を取得
	channelDetails, err := s.ytService.FetchChannelDetail(ctx, channelIDs)
	if err != nil {
		return nil, false, nil, err
	}

	fetchedAt := time.Now().UTC()
	q := sqlc.New(s.db)

	// チャンネルを保存
	savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
	for _, cd := range channelDetails {
		ch, err := channel.NewChannel(fetchedAt, fetchedAt, cd)
		if err != nil {
			slog.Info("failed to NewChannel(search)", "error", err)
			continue
		}
		if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
			slog.Info("failed to saveChannel(search)", "error", err)
			continue
		}
		savedChannels[cd.ID] = ch.ID
	}

	// 動画を保存し、結果を構築（検索順序を維持）
	items := make([]SearchVideoView, 0, len(videoIDs))
	for _, vid := range videoIDs {
		vd, ok := videoDetails[vid]
		if !ok {
			continue
		}
		channelUUID, ok := savedChannels[vd.ChannelID]
		if !ok {
			continue
		}
		cd, ok := channelDetails[vd.ChannelID]
		if !ok {
			continue
		}

		v, err := video.NewVideo(channelUUID, fetchedAt, vd)
		if err != nil {
			slog.Info("failed to NewVideo(search)", "error", err)
			continue
		}
		if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
			slog.Info("failed to saveVideo(search)", "error", err)
			continue
		}

		items = append(items, SearchVideoView{
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
	}

	// ページネーション
	hasNext := nextPageToken != ""
	var nextCursor *string
	if hasNext {
		nextCursor = &nextPageToken
	}

	// limitで切り詰め
	if len(items) > limit {
		items = items[:limit]
	}

	return items, hasNext, nextCursor, nil
}
