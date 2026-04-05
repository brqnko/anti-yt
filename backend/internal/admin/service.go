package admin

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool

	ytService youtube_d.Service
}

func NewService(db *pgxpool.Pool, ytService youtube_d.Service) *Service {
	return &Service{
		db:        db,
		ytService: ytService,
	}
}

func (s *Service) CreateNewValuableChannel(ctx context.Context, externalChannelID string, reason, description string) (_ *channel.ValuableChannel, err error) {
	defer util.Wrap(&err, "admin.(*Service).CreateNewValuableChannel")

	// ユーザーはURLやハンドルやチャンネルIDで入力してくる
	channelIDOrHandle, err := youtube_d.ExtractChannelIDOrHandle(externalChannelID)
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
	foundChannel, err := channel.NewChannelRepository(q).FindByIdOrHandle(ctx, channelIDOrHandle)
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, core.ErrNotFound) {
		// YouTubeからチャンネル情報を取得
		channelDetail, err := s.ytService.FetchChannelDetailByIDOrHandle(ctx, channelIDOrHandle)
		fetchedAt := time.Now().UTC()
		if err != nil {
			return nil, err
		}

		// YouTubeで取得したチャンネル情報をシステムのエンティティに変換
		ch, err := channel.NewChannel(fetchedAt, fetchedAt, channelDetail)
		if err != nil {
			return nil, err
		}

		// チャンネルを保存する
		if _, err := channel.NewChannelRepository(sqlc.New(s.db)).Save(ctx, ch); err != nil {
			return nil, err
		}

		// チャンネルの投稿動画(IDのみ)をAPIから取得する
		uploadIDs, _, err := s.ytService.FetchPlaylistVideoIDs(ctx, string(ch.Channel.UploadsPlaylistID), "")
		if err != nil {
			return nil, err
		}

		// チャンネルの投稿動画IDリストから、それぞれの動画情報を取得する
		videoDetails, err := s.ytService.FetchVideoDetail(ctx, uploadIDs)
		if err != nil {
			return nil, err
		}

		// 取得した情報をDBに保存する
		for _, vd := range videoDetails {
			v, err := video.NewVideo(ch.ID, fetchedAt, vd)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(create valuable channel)", slog.Any("error", err))
				continue
			}

			if _, err := video.NewVideoRepository(sqlc.New(s.db)).Save(ctx, v); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(create valuable channel)", slog.Any("error", err))
			}
		}

		foundChannel = ch
	}

	// ValuableChannelドメインオブジェクトを作成
	vc, err := channel.NewValuableChannel(foundChannel.ID, reason, description)
	if err != nil {
		return nil, err
	}

	// ValuableChannelを保存
	if _, err := channel.NewValuableChannelRepository(q).Save(ctx, vc); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return vc, nil
}

func (s *Service) UpdateValuableChannel(ctx context.Context, externalChannelID string, reaason, description *string) (_ *channel.ValuableChannel, err error) {
	defer util.Wrap(&err, "admin.(*Service).UpdateValuableChannel")

	// ユーザーはURLやハンドルやチャンネルIDで入力してくる
	channelIDOrHandle, err := youtube_d.ExtractChannelIDOrHandle(externalChannelID)
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

	// チャンネルを検索
	foundChannel, err := channel.NewChannelRepository(q).FindByIdOrHandle(ctx, channelIDOrHandle)
	if err != nil {
		return nil, err
	}

	// ValuableChannelをロッキングリードで取得
	vc, err := channel.NewValuableChannelRepository(q).FindForUpdate(ctx, foundChannel.ID)
	if err != nil {
		return nil, err
	}

	// 部分更新
	if err := vc.SetReasonCode(reaason); err != nil {
		return nil, err
	}
	if err := vc.SetDescription(description); err != nil {
		return nil, err
	}

	// 保存
	if _, err := channel.NewValuableChannelRepository(q).Save(ctx, vc); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return vc, nil
}

func (s *Service) RemoveValuableChannel(ctx context.Context, externalChannelID string) (err error) {
	defer util.Wrap(&err, "admin.(*Service).RemoveValuableChannel")

	channelIDOrHandle, err := youtube_d.ExtractChannelIDOrHandle(externalChannelID)
	if err != nil {
		return err
	}

	q := sqlc.New(s.db)

	foundChannel, err := channel.NewChannelRepository(q).FindByIdOrHandle(ctx, channelIDOrHandle)
	if err != nil {
		return err
	}

	if err := channel.NewValuableChannelRepository(q).Remove(ctx, foundChannel.ID); err != nil {
		return err
	}

	return nil
}
