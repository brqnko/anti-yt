package channel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
	ytService youtube_d.YouTubeAPIService

	rssFetchDuration time.Duration
}

var (
	ErrInvalidSubscriptionLimit = errors.New("invalid subscription limit: out of range (should be [1..50])")
	ErrInvalidGetUploadLimit    = errors.New("invalid get upload limit: out of range (should be [1..50])")
	ErrSubscriptionNotFound     = errors.New("subscription not found")
)

func NewService(
	db *pgxpool.Pool,
	ytService youtube_d.YouTubeAPIService,
	rssFetchDuration time.Duration,
) (*Service, error) {
	return &Service{
		db:               db,
		ytService:        ytService,
		rssFetchDuration: rssFetchDuration,
	}, nil
}

func (s *Service) SubscribeChannel(ctx context.Context, channelText string) (*SubscribedChannel, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	channelID, err := util.ExtractChannelIDOrHandle(channelText)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()
	q := sqlc.New(tx)

	if err := q.AcquireAdvisoryXactLock(ctx, util.Sha256Int64([]byte(channelID))); err != nil {
		return nil, fmt.Errorf("acquireAdvisoryXactLock: %w", err)
	}
	found, err := q.GetChannelByIdOrHandle(ctx, sqlc.GetChannelByIdOrHandleParams{
		ExternalID:       channelID,
		ExternalCustomID: channelID,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getChannelByIdOrHandle: %w", err)
		}

		info, err := s.ytService.FetchChannelInfo(ctx, channelID)
		if err != nil {
			return nil, fmt.Errorf("fetchChannelInfo: %w", err)
		}

		// NOTE: fetchの結果をキャッシュするため、トランザクションの外で行う
		saved, err := sqlc.New(s.db).SaveChannel(ctx, sqlc.SaveChannelParams{
			ExternalID:               info.ID,
			ExternalDisplayName:      info.DisplayName,
			ExternalCustomID:         info.CustomID,
			ExternalIconUrl:          info.IconURL,
			ExternalDescription:      info.Description,
			ExternalSubscribersCount: int64(info.SubscribersCount),
			ExternalCreatedAt:        info.CreatedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("saveChannel: %w", err)
		}
		found = sqlc.GetChannelByIdOrHandleRow{
			PublicID:                 saved.PublicID,
			ExternalID:               info.ID,
			ExternalCustomID:         info.CustomID,
			ExternalDescription:      info.Description,
			ExternalIconUrl:          info.IconURL,
			ExternalSubscribersCount: int64(info.SubscribersCount),
			ExternalCreatedAt:        info.CreatedAt,
			MChannelID:               saved.MChannelID,
			ExternalDisplayName:      info.DisplayName,
		}
	}

	saveChannelSubscription, err := q.SaveChannelSubscription(ctx, sqlc.SaveChannelSubscriptionParams{
		ChannelID:    found.MChannelID,
		UserPublicID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("saveChannelSubscription: %w", err)
	}

	subscribed, err := NewSubscribedChannel(
		saveChannelSubscription.PublicID,
		found.PublicID,
		saveChannelSubscription.CreatedAt,
		found.ExternalID,
		found.ExternalDisplayName,
		found.ExternalCustomID,
		found.ExternalDescription,
		found.ExternalIconUrl,
		int(found.ExternalSubscribersCount),
		found.ExternalCreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return subscribed, nil
}

func (s *Service) UnsubscribeChannel(ctx context.Context, subscriptionID uuid.UUID) error {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	q := sqlc.New(s.db)
	rowsAffected, err := q.DeleteChannelSubscription(ctx, sqlc.DeleteChannelSubscriptionParams{
		SubscriptionPublicID: subscriptionID,
		UserPublicID:         userID,
	})
	if err != nil {
		return fmt.Errorf("deleteChannelSubscription: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

// cursorは最後の登録チャンネルのPublicID
func (s *Service) GetSubscriptions(ctx context.Context, limit int, cursor *uuid.UUID) (channels []*SubscribedChannel, hasNext bool, err error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, false, err
	}

	if limit < 1 || 50 < limit {
		return []*SubscribedChannel{}, false, ErrInvalidSubscriptionLimit
	}

	q := sqlc.New(s.db)
	subscriptions, err := q.GetChannelSubscriptions(ctx, sqlc.GetChannelSubscriptionsParams{
		UserPublicID:   userID,
		CursorPublicID: cursor,
		QueryLimit:     int32(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("getChannelSubscriptions: %w", err)
	}
	res := make([]*SubscribedChannel, min(len(subscriptions), limit))
	for i, subscription := range subscriptions {
		if i >= limit {
			break
		}
		s, err := NewSubscribedChannel(
			subscription.PublicID,
			subscription.ChannelPublicID,
			subscription.CreatedAt,
			subscription.ExternalID,
			subscription.ExternalDisplayName,
			subscription.ExternalCustomID,
			subscription.ExternalDescription,
			subscription.ExternalIconUrl,
			int(subscription.ExternalSubscribersCount),
			subscription.ExternalCreatedAt,
		)
		if err != nil {
			return nil, false, err
		}
		res[i] = s
	}

	return res, len(subscriptions) == limit+1, nil
}

func (s *Service) GetChannelUploads(ctx context.Context, channelID uuid.UUID, cursor *uuid.UUID, limit int) (videos []*video.Video, hasNext bool, err error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, false, err
	}
	if limit < 1 || 50 < limit {
		return nil, false, ErrInvalidGetUploadLimit
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()
	q := sqlc.New(tx)

	rssFetchedAt, err := q.GetChannelRSSFetchedAtForUpdate(ctx, channelID)
	if err != nil {
		return nil, false, fmt.Errorf("getChannelRSSFetchedAtForUpdate: %w", err)
	}
	if time.Now().UTC().Sub(rssFetchedAt.RssFetchedAt) > s.rssFetchDuration {
		rss, err := s.ytService.FetchRSSFeed(ctx, rssFetchedAt.ExternalID)
		if err != nil {
			return nil, false, fmt.Errorf("fetchRSSFeed: %w", err)
		}

		videoIDs := make([]string, len(rss))
		for i, v := range rss {
			videoIDs[i] = v.VideoID
		}
		videoInfoMap, err := s.ytService.FetchVideoInfo(ctx, videoIDs)
		if err != nil {
			return nil, false, fmt.Errorf("fetchVideoInfo: %w", err)
		}

		now := time.Now().UTC()
		// キャッシュはトランザクションでロールバックしたくない
		for _, v := range rss {
			videoInfo, ok := videoInfoMap[v.VideoID]
			if !ok {
				continue
			}

			if err := sqlc.New(s.db).SaveVideo(ctx, sqlc.SaveVideoParams{
				MChannelID:            rssFetchedAt.MChannelID,
				ExternalID:            v.VideoID,
				ExternalTitle:         v.Title,
				ExternalDescription:   v.Description,
				FetchedAt:             now,
				ExternalCreatedAt:     v.CreatedAt,
				ExternalThumbnailUrl:  v.ThumbnailURL,
				ExternalLengthSeconds: videoInfo.LengthSeconds,
			}); err != nil {
				return nil, false, fmt.Errorf("saveVideo: %w", err)
			}
		}

		if _, err := q.MarkChannelRSSAsFetched(ctx, []int64{rssFetchedAt.MChannelID}); err != nil {
			return nil, false, fmt.Errorf("markChannelRSSAsFetched: %w", err)
		}
	}

	getChannelVideos, err := q.GetChannelVideos(ctx, sqlc.GetChannelVideosParams{
		UserID:     userID,
		ChannelID:  channelID,
		Cursor:     cursor,
		QueryLimit: (int32)(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("getChannelVideos: %w", err)
	}

	videosToReturn := make([]*video.Video, min(limit, len(getChannelVideos)))
	for i, getChannelVideo := range getChannelVideos {
		if i >= limit {
			break
		}
		// NOTE: pgxはNULLが帰ってきても、0として解釈する。
		// 0ならnilとして扱うように、コンストラクタ側で処理してます。
		v := video.NewVideo(
			getChannelVideo.PublicID,
			getChannelVideo.ChannelID,
			getChannelVideo.ExternalThumbnailUrl,
			getChannelVideo.ExternalChannelIconUrl,
			getChannelVideo.ExternalTitle,
			getChannelVideo.ExternalChannelDisplayname,
			getChannelVideo.ExternalCreatedAt,
			getChannelVideo.ExternalLengthSeconds,
			getChannelVideo.LastWatchSeconds,
		)
		videosToReturn[i] = v
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit: %w", err)
	}

	return videosToReturn, len(getChannelVideos) == limit+1, nil
}

func (s *Service) GetFeed(ctx context.Context, cursor *uuid.UUID, limit int) (videos []*video.Video, hasNext bool, err error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, false, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Warn("failed to rollback transaction", "error", err)
		}
	}()
	q := sqlc.New(tx)

	forUpdates, err := q.GetChannelsToFetchRSSForUpdate(ctx, sqlc.GetChannelsToFetchRSSForUpdateParams{
		UserID:   userID,
		RssFetch: time.Now().UTC().Add(-s.rssFetchDuration),
	})
	if err != nil {
		return nil, false, fmt.Errorf("getChannelsToFetchRSSForUpdate: %w", err)
	}

	if len(forUpdates) != 0 {
		for _, forUpdate := range forUpdates {
			feed, err := s.ytService.FetchRSSFeed(ctx, forUpdate.ExternalID)
			// TODO: チャンネルが削除されてrssの取得に失敗した場合のケースを考慮する
			if err != nil {
				return nil, false, fmt.Errorf("fetchRSSFeed: %w", err)
			}

			videoIDs := make([]string, len(feed))
			for i, f := range feed {
				videoIDs[i] = f.VideoID
			}
			// TODO: 現在はforUpdates分YouTube Data APIにリクエストを投げているが、一括で取得したい
			videoInfoMap, err := s.ytService.FetchVideoInfo(ctx, videoIDs)
			if err != nil {
				return nil, false, fmt.Errorf("fetchVideoInfo: %w", err)
			}

			now := time.Now().UTC()
			for _, f := range feed {
				videoInfo, ok := videoInfoMap[f.VideoID]
				// TODO: 削除された動画どうするか考える
				if !ok {
					continue
				}

				if err := sqlc.New(s.db).SaveVideo(ctx, sqlc.SaveVideoParams{
					MChannelID:            forUpdate.MChannelID,
					ExternalID:            f.VideoID,
					ExternalTitle:         f.Title,
					ExternalDescription:   f.Description,
					FetchedAt:             now,
					ExternalCreatedAt:     f.CreatedAt,
					ExternalThumbnailUrl:  f.ThumbnailURL,
					ExternalLengthSeconds: videoInfo.LengthSeconds,
				}); err != nil {
					return nil, false, fmt.Errorf("saveVideo: %w", err)
				}
			}
		}

		channelIDs := make([]int64, len(forUpdates))
		for i, u := range forUpdates {
			channelIDs[i] = u.MChannelID
		}
		if _, err := q.MarkChannelRSSAsFetched(ctx, channelIDs); err != nil {
			return nil, false, fmt.Errorf("markChannelRSSAsFetched: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit: %w", err)
	}

	res, err := sqlc.New(s.db).GetSubscribingChannelFeed(ctx, sqlc.GetSubscribingChannelFeedParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: (int32)(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("getSubscribingChannelFeed: %w", err)
	}

	videosToReturn := make([]*video.Video, 0, min(len(res), limit))
	for i, r := range res {
		if i >= limit {
			break
		}
		v := video.NewVideo(
			r.VideoID,
			r.ChannelID,
			r.ExternalVideoThumbnailUrl,
			r.ExternalChannelIconUrl,
			r.ExternalTitle,
			r.ExternalDisplayname,
			r.ExternalCreatedAt,
			r.ExternalLengthSeconds,
			r.LastWatchSeconds,
		)
		videosToReturn = append(videosToReturn, v)
	}

	return videosToReturn, len(res) == limit+1, nil
}
