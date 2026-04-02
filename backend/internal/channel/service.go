package channel

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
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
	db        *pgxpool.Pool
	ytService youtube_d.Service

	channelQS         ChannelQueryService
	valuableChannelQS ValuableChannelQueryService

	rssFetchDuration time.Duration
}

var (
	ErrInvalidSubscriptionLimit = core.NewDomainError("channel.invalid_subscription_limit", "invalid subscription limit: out of range (should be [1..50])")
	ErrInvalidGetUploadLimit    = core.NewDomainError("channel.invalid_get_upload_limit", "invalid get upload limit: out of range (should be [1..50])")
	ErrInvalidChannelID         = core.NewDomainError("channel.invalid_channel_id", "invalid channel id")
)

func NewService(
	db *pgxpool.Pool,
	ytService youtube_d.Service,
	rssFetchDuration time.Duration,
) *Service {
	return &Service{
		db:                db,
		ytService:         ytService,
		rssFetchDuration:  rssFetchDuration,
		channelQS:         NewChannelQueryService(db),
		valuableChannelQS: NewValuableChannelQueryService(db),
	}
}

func (s *Service) SubscribeChannel(ctx context.Context, userID uuid.UUID, channelText string) (_ *Channel, err error) {
	defer util.Wrap(&err, "Service.SubscribeChannel")

	// ユーザーはURLやハンドルやチャンネルIDで入力してくる
	channelIDOrHandle, err := youtube_d.ExtractChannelIDOrHandle(channelText)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback transaction", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	if err := database_d.TryAdLock(ctx, q, []byte(channelIDOrHandle)); err != nil {
		return nil, err
	}

	// すでに保存されているかを確認する
	// 保存されてない場合はfetchしてそれを使う
	foundChannel, err := NewChannelRepository(q).FindByIdOrHandle(ctx, channelIDOrHandle)
	if err != nil && !errors.Is(err, core.ErrNotFound) { // ただのDBエラー
		return nil, err
	}
	if errors.Is(err, core.ErrNotFound) { // 保存されてない場合
		// YouTubeからチャンネル情報を取得
		channelDetail, err := s.ytService.FetchChannelDetailByIDOrHandle(ctx, channelIDOrHandle)
		fetchedAt := time.Now().UTC()
		if err != nil {
			return nil, err
		}

		// YouTubeで取得したチャンネル情報をシステムのエンティティに変換
		channel, err := NewChannel(fetchedAt, fetchedAt, channelDetail)
		if err != nil {
			return nil, err
		}

		// チャンネルを保存する
		// NOTE: fetchの結果をキャッシュするため、トランザクションの外で行う
		_, err = NewChannelRepository(sqlc.New(s.db)).Save(ctx, channel)
		if err != nil {
			return nil, err
		}

		// チャンネルの投稿動画(IDのみ)をAPIから取得する
		// NOTE: 新しいチャンネルは、新規ユーザーによる可能性が高い. 新規ユーザーは貴重なため、RSS Feedより確実なYouTube Data APIで動画を取得しておく
		// ちなみにYouTubeのPlaylistのIDには種類がある(ユーザーのアップロードしたリストや普通のプレイリスト、自動生成されたプレイリストなど)
		// FetchPlaylistVideoIDsは汎用的なメソッドのためstirngを受け取っている
		uploadIDs, _, err := s.ytService.FetchPlaylistVideoIDs(ctx, string(channel.Channel.UploadsPlaylistID), "")
		if err != nil {
			return nil, err
		}

		// チャンネルの投稿動画IDリストから、それぞれの動画情報を取得する
		videoDetails, err := s.ytService.FetchVideoDetail(ctx, uploadIDs)
		if err != nil {
			return nil, err
		}

		// 取得した情報をDBに保存する. キャッシュのためトランザクション外で行う
		for _, vd := range videoDetails {
			v, err := video.NewVideo(channel.ID, fetchedAt, vd)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to newVideo", slog.Any("error", err))
				continue
			}

			if _, err := video.NewVideoRepository(sqlc.New(s.db)).Save(ctx, v); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video", slog.Any("error", err))
			}
		}

		foundChannel = channel
	}

	subscribedChannel, err := NewSubscribedChannel(foundChannel.ID, userID)
	if err != nil {
		return nil, err
	}

	if _, err := NewChannelRepository(q).SaveSubscription(ctx, subscribedChannel); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return foundChannel, nil
}

func (s *Service) UnsubscribeChannel(ctx context.Context, userID, channelID uuid.UUID) (err error) {
	defer util.Wrap(&err, "Service.UnsubscribeChannel")

	_, err = NewChannelRepository(sqlc.New(s.db)).RemoveSubscription(ctx, userID, channelID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetSubscriptions(ctx context.Context, userID uuid.UUID, limit int32, cursor *uuid.UUID) (_ []GetSubscriptionsView, _ bool, err error) {
	defer util.Wrap(&err, "Service.GetSubscriptions")

	if limit < 1 || 50 < limit { // openapiもあるが一応チェック
		return nil, false, ErrInvalidSubscriptionLimit
	}

	channels, err := s.channelQS.GetSubscriptions(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(channels) > int(limit) { // NOTE: int -> int32の変換よりもint32 -> intの方が安全
		return channels[:limit], true, nil
	}
	return channels, false, nil
}

func (s *Service) GetChannelUploads(ctx context.Context, userID, channelID uuid.UUID, cursor *uuid.UUID, limit int32, order string) (_ []GetChannelUploadsView, _ bool, err error) {
	defer util.Wrap(&err, "Service.GetChannelUploads")

	if limit < 1 || 50 < limit {
		return nil, false, ErrInvalidGetUploadLimit
	}

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

	// ロッキングリード
	lockedChannel, err := NewChannelRepository(q).FindForUpdate(ctx, channelID)
	if err != nil {
		return nil, false, err
	}
	if lockedChannel.ShouldFetchRSSFeed(s.rssFetchDuration) {
		// PlaylistAPIから動画ID一覧を取得する
		videoIDs, _, err := s.ytService.FetchPlaylistVideoIDs(ctx, string(lockedChannel.Channel.UploadsPlaylistID), "")
		if err != nil {
			return nil, false, err
		}

		// 動画ID一覧から動画の詳細情報を取得する
		videoDetailMap, err := s.ytService.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			return nil, false, err
		}

		fetchedAt := time.Now().UTC()
		for _, videoDetail := range videoDetailMap {
			v, err := video.NewVideo(lockedChannel.ID, fetchedAt, videoDetail)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video", slog.Any("error", err))
				continue
			}

			if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save new video", slog.Any("error", err))
				continue
			}
		}

		lockedChannel.MarkAsRSSFetched()
		if _, err := NewChannelRepository(q).Save(ctx, lockedChannel); err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	videos, err := s.channelQS.GetChannelUploads(ctx, userID, channelID, cursor, limit+1, order)
	if err != nil {
		return nil, false, err
	}

	if len(videos) > int(limit) {
		return videos[:limit], true, nil
	}
	return videos, false, nil
}

func (s *Service) GetChannelDetail(ctx context.Context, channelID uuid.UUID) (_ GetChannelDetailView, err error) {
	defer util.Wrap(&err, "Service.GetChannelDetail")

	detail, err := s.channelQS.GetChannelDetail(ctx, channelID)
	if err != nil {
		return GetChannelDetailView{}, err
	}

	return detail, nil
}

func (s *Service) GetChannelFeeds(ctx context.Context) (_ []GetValuableChannelView, err error) {
	defer util.Wrap(&err, "Service.GetChannelFeeds")

	channels, err := s.valuableChannelQS.GetValuableChannels(ctx)
	if err != nil {
		return nil, err
	}

	return channels, nil
}
