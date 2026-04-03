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
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
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
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to newVideo", slog.Any("error", err))
				continue
			}

			if _, err := video.NewVideoRepository(sqlc.New(s.db)).Save(ctx, v); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video", slog.Any("error", err))
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

func (s *Service) ImportChannelVideos(ctx context.Context, externalChannelID string) (_ int, err error) {
	defer util.Wrap(&err, "admin.(*Service).ImportChannelVideos")

	channelIDOrHandle, err := youtube_d.ExtractChannelIDOrHandle(externalChannelID)
	if err != nil {
		return 0, err
	}

	q := sqlc.New(s.db)

	foundChannel, err := channel.NewChannelRepository(q).FindByIdOrHandle(ctx, channelIDOrHandle)
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return 0, err
	}
	if errors.Is(err, core.ErrNotFound) {
		channelDetail, err := s.ytService.FetchChannelDetailByIDOrHandle(ctx, channelIDOrHandle)
		fetchedAt := time.Now().UTC()
		if err != nil {
			return 0, err
		}

		ch, err := channel.NewChannel(fetchedAt, fetchedAt, channelDetail)
		if err != nil {
			return 0, err
		}

		if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
			return 0, err
		}

		foundChannel = ch
	}

	fetchedAt := time.Now().UTC()
	savedCount := 0
	pageToken := ""
	for {
		videoIDs, nextPageToken, err := s.ytService.FetchPlaylistVideoIDs(ctx, string(foundChannel.Channel.UploadsPlaylistID), pageToken)
		if err != nil {
			return savedCount, err
		}

		if len(videoIDs) > 0 {
			videoDetails, err := s.ytService.FetchVideoDetail(ctx, videoIDs)
			if err != nil {
				util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to fetch video detail", slog.Any("error", err))
			} else {
				for _, vd := range videoDetails {
					v, err := video.NewVideo(foundChannel.ID, fetchedAt, vd)
					if err != nil {
						util.LoggerFromContext(ctx).InfoContext(ctx, "failed to newVideo", slog.Any("error", err))
						continue
					}
					if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
						util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video", slog.Any("error", err))
						continue
					}
					savedCount++
				}
			}
		}

		if nextPageToken == "" {
			break
		}
		pageToken = nextPageToken
	}

	return savedCount, nil
}

func (s *Service) ImportChannelPlaylists(ctx context.Context, externalChannelID string) (_ int, err error) {
	defer util.Wrap(&err, "admin.(*Service).ImportChannelPlaylists")

	channelIDOrHandle, err := youtube_d.ExtractChannelIDOrHandle(externalChannelID)
	if err != nil {
		return 0, err
	}

	q := sqlc.New(s.db)

	foundChannel, err := channel.NewChannelRepository(q).FindByIdOrHandle(ctx, channelIDOrHandle)
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return 0, err
	}
	if errors.Is(err, core.ErrNotFound) {
		channelDetail, err := s.ytService.FetchChannelDetailByIDOrHandle(ctx, channelIDOrHandle)
		if err != nil {
			return 0, err
		}
		fetchedAt := time.Now().UTC()

		ch, err := channel.NewChannel(fetchedAt, fetchedAt, channelDetail)
		if err != nil {
			return 0, err
		}

		if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
			return 0, err
		}

		foundChannel = ch
	}

	var ytPlaylists []youtube_d.Playlist
	pageToken := ""
	for {
		playlists, nextPageToken, err := s.ytService.FetchChannelPlaylists(ctx, foundChannel.Channel.ID, pageToken)
		if err != nil {
			return 0, err
		}
		ytPlaylists = append(ytPlaylists, playlists...)

		if nextPageToken == "" {
			break
		}
		pageToken = nextPageToken
	}

	savedPlaylistCount := 0

	for _, ytPlaylist := range ytPlaylists {
		util.LoggerFromContext(ctx).InfoContext(ctx, "importing playlist", slog.String("playlistID", ytPlaylist.ID), slog.String("title", ytPlaylist.Title))

		// プレイリストの動画ID一覧を全ページ取得
		var allVideoIDs []youtube_d.VideoID
		videoPageToken := ""
		for {
			videoIDs, nextVideoPageToken, err := s.ytService.FetchPlaylistVideoIDs(ctx, ytPlaylist.ID, videoPageToken)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch playlist video IDs", slog.Any("error", err))
				break
			}
			allVideoIDs = append(allVideoIDs, videoIDs...)

			if nextVideoPageToken == "" {
				break
			}
			videoPageToken = nextVideoPageToken
		}

		// 動画詳細を取得(mapに存在しないものだけリクエスト)
		videoDetailMap := make(map[youtube_d.VideoID]youtube_d.Video)
		var toRequestVideos []youtube_d.VideoID
		for _, vid := range allVideoIDs {
			if _, ok := videoDetailMap[vid]; ok {
				continue
			}
			toRequestVideos = append(toRequestVideos, vid)
			if len(toRequestVideos) >= 50 {
				videoDetails, err := s.ytService.FetchVideoDetail(ctx, toRequestVideos)
				if err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch video details", slog.Any("error", err))
				} else {
					for id, vd := range videoDetails {
						videoDetailMap[id] = vd
					}
				}
				toRequestVideos = nil
			}
		}
		if len(toRequestVideos) > 0 {
			videoDetails, err := s.ytService.FetchVideoDetail(ctx, toRequestVideos)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch video details", slog.Any("error", err))
			} else {
				for id, vd := range videoDetails {
					videoDetailMap[id] = vd
				}
			}
		}

		// チャンネル詳細を取得(mapに存在しないものだけリクエスト)
		channelDetailMap := make(map[youtube_d.ChannelID]youtube_d.Channel)
		var toRequestChannels []youtube_d.ChannelID
		for _, vd := range videoDetailMap {
			if _, ok := channelDetailMap[vd.ChannelID]; ok {
				continue
			}
			toRequestChannels = append(toRequestChannels, vd.ChannelID)
			if len(toRequestChannels) >= 50 {
				channelDetails, err := s.ytService.FetchChannelDetail(ctx, toRequestChannels)
				if err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch channel details", slog.Any("error", err))
				} else {
					for id, cd := range channelDetails {
						channelDetailMap[id] = cd
					}
				}
				toRequestChannels = nil
			}
		}
		if len(toRequestChannels) > 0 {
			channelDetails, err := s.ytService.FetchChannelDetail(ctx, toRequestChannels)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch channel details", slog.Any("error", err))
			} else {
				for id, cd := range channelDetails {
					channelDetailMap[id] = cd
				}
			}
		}

		fetchedAt := time.Now().UTC()

		// チャンネルを保存
		savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
		for _, cd := range channelDetailMap {
			ch, err := channel.NewChannel(fetchedAt, fetchedAt, cd)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new channel", slog.Any("error", err))
				continue
			}
			if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save channel", slog.Any("error", err))
				continue
			}
			savedChannels[cd.ID] = ch.ID
		}

		// 動画を保存
		savedVideos := make(map[youtube_d.VideoID]int64)
		for _, vd := range videoDetailMap {
			channelUUID, ok := savedChannels[vd.ChannelID]
			if !ok {
				continue
			}
			v, err := video.NewVideo(channelUUID, fetchedAt, vd)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to newVideo", slog.Any("error", err))
				continue
			}
			savedVideoID, err := video.NewVideoRepository(q).Save(ctx, v)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video", slog.Any("error", err))
				continue
			}
			savedVideos[v.Video.ID] = savedVideoID
		}

		// プレイリストを作成して動画を挿入
		pl, err := playlist.NewPlaylist(
			ytPlaylist.Title,
			ytPlaylist.Description,
			"public",
			"external_auto",
			playlist.WithPlaylistRegisteredAt(ytPlaylist.CreatedAt),
			playlist.WithPlaylistChannelID(foundChannel.ID),
		)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to create playlist domain", slog.Any("error", err))
			continue
		}

		playlistRow, err := playlist.NewPlaylistRepository(q).Save(ctx, pl)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save playlist", slog.Any("error", err))
			continue
		}

		var videoInternalIDs []int64
		for _, vid := range allVideoIDs {
			savedVideoID, ok := savedVideos[vid]
			if !ok {
				continue
			}
			videoInternalIDs = append(videoInternalIDs, savedVideoID)
		}

		if len(videoInternalIDs) > 0 {
			if err := playlist.NewPlaylistRepository(q).BulkInsertVideos(ctx, playlistRow, videoInternalIDs); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to bulk insert videos", slog.Any("error", err))
				continue
			}
		}

		if err := pl.SetVideoCount(len(videoInternalIDs)); err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to set video count", slog.Any("error", err))
			continue
		}

		if _, err := playlist.NewPlaylistRepository(q).Save(ctx, pl); err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save playlist with video count", slog.Any("error", err))
			continue
		}

		savedPlaylistCount++
	}

	return savedPlaylistCount, nil
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
