package playlist

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
	playlistQS    PlaylistQueryService
}

func NewService(db *pgxpool.Pool, youtubeClient youtube_d.Client, feedRepo database_d.FeedRepository) *Service {
	return new(Service{
		db:            db,
		youtubeClient: youtubeClient,
		feedRepo:      feedRepo,
		playlistQS:    NewPlaylistQueryService(db),
	})
}

func (s *Service) CreatePlaylist(ctx context.Context, userID uuid.UUID, title, description, visibilityStr, playlistTypeStr string, basePlaylistUrl *string) (_ *Playlist, err error) {
	defer util.Wrap(&err, "playlist.(*Service).CreatePlaylist(userID=%s)", userID)

	playlist, err := NewPlaylist(
		userID,
		title,
		description,
		visibilityStr,
		playlistTypeStr,
	)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	count, err := q.CountUserPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 20 {
		return nil, ErrTooManyPlaylists
	}

	// NOTE: insertした内容はcommitするまで他のトランザクションから見れない(repeatable read)ので、ロックなどをつける意味はない
	row, err := NewPlaylistRepository(q).Save(ctx, playlist)
	if err != nil {
		return nil, err
	}

	if basePlaylistUrl == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return playlist, nil
	}

	// YouTubeからプレイリストをimportする
	playlistID, err := youtube_d.ExtractPlaylistID(*basePlaylistUrl)
	if err != nil {
		return nil, err
	}

	var allVideoIDs []int64 // 順序を保持するため
	var nextPageToken string
	for {
		// プレイリストから動画IDのリストを取得
		videoIDs, pageToken, err := s.youtubeClient.FetchPlaylistVideoIDs(ctx, playlistID, nextPageToken)
		if err != nil {
			return nil, err
		}

		// 動画IDのリストから詳細を取得
		videoDetails, err := s.youtubeClient.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			return nil, err
		}

		// チャンネル詳細を取得
		channelIDs := make([]youtube_d.ChannelID, len(videoDetails))
		i := 0
		for _, vd := range videoDetails {
			channelIDs[i] = vd.ChannelID
			i++
		}
		channelDetails, err := s.youtubeClient.FetchChannelDetail(ctx, channelIDs)
		if err != nil {
			return nil, err
		}

		fetchedAt := time.Now().UTC()

		// NOTE: チャンネルを先に保存するので、動画があってチャンネルがないようなことは発生しないため、トランザクションは使用しない
		// チャンネルの詳細情報を先に保存する
		savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
		for _, channelDetail := range channelDetails {
			ch, err := channel.NewChannel(fetchedAt, fetchedAt.AddDate(-1, 0, 0), channelDetail)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new channel(create playlist)", slog.Any("error", err))
				continue
			}

			if _, err := channel.NewChannelRepository(sqlc.New(s.db)).Save(ctx, ch); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save channel(create playlist)", slog.Any("error", err))
				continue
			}
			savedChannels[channelDetail.ID] = ch.ID
		}

		// 動画の詳細情報を保存する
		// 動画の保存自体に順番は関係ない
		videoIDToint64 := make(map[youtube_d.VideoID]int64)
		for _, videoDetail := range videoDetails {
			channelUUID, ok := savedChannels[videoDetail.ChannelID]
			if !ok {
				util.LoggerFromContext(ctx).InfoContext(ctx, "channel not found in saved channels(create playlist)", slog.String("channelID", string(videoDetail.ChannelID)))
				continue
			}
			v, err := video.NewVideo(channelUUID, fetchedAt, videoDetail)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(create playlist)", slog.Any("error", err))
				continue
			}

			q := sqlc.New(s.db)
			savedVideoID, err := video.NewVideoRepository(q).Save(ctx, v)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(create playlist)", slog.Any("error", err))
				continue
			}
			if err := s.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
				util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(create playlist)", slog.Any("error", err))
			}
			videoIDToint64[v.Video.ID] = savedVideoID
		}

		for _, vid := range videoIDs {
			savedVideoID, ok := videoIDToint64[vid]
			if !ok {
				util.LoggerFromContext(ctx).InfoContext(ctx, "video not found in saved videos(create playlist)", slog.String("videoID", string(vid)))
				continue
			}
			allVideoIDs = append(allVideoIDs, savedVideoID)
		}

		if pageToken == "" {
			break
		}
		nextPageToken = pageToken
	}

	if err := NewPlaylistRepository(q).BulkInsertVideos(ctx, row, allVideoIDs); err != nil {
		return nil, err
	}

	if err := playlist.SetVideoCount(len(allVideoIDs)); err != nil {
		return nil, err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *Service) CreatePlaylistWithOAuthClient(ctx context.Context, userID uuid.UUID, title, description, visibilityStr, playlistTypeStr string, oauthClient *youtube_d.OAuthClient, playlistID string) (_ *Playlist, err error) {
	defer util.Wrap(&err, "playlist.(*Service).CreatePlaylistWithOAuthClient(userID=%s)", userID)

	playlist, err := NewPlaylist(
		userID,
		title,
		description,
		visibilityStr,
		playlistTypeStr,
	)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	count, err := q.CountUserPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 20 {
		return nil, ErrTooManyPlaylists
	}

	row, err := NewPlaylistRepository(q).Save(ctx, playlist)
	if err != nil {
		return nil, err
	}

	var allVideoIDs []int64
	var nextPageToken string
	for {
		videoIDs, pageToken, err := oauthClient.FetchPlaylistVideoIDs(ctx, playlistID, nextPageToken)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch playlist video ids(create playlist with access token)", slog.Any("error", err))
			break
		}

		videoDetails, err := s.youtubeClient.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch video detail(create playlist with access token)", slog.Any("error", err))
			break
		}

		channelIDs := make([]youtube_d.ChannelID, 0, len(videoDetails))
		for _, vd := range videoDetails {
			channelIDs = append(channelIDs, vd.ChannelID)
		}
		channelDetails, err := s.youtubeClient.FetchChannelDetail(ctx, channelIDs)
		if err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch channel detail(create playlist with access token)", slog.Any("error", err))
			break
		}

		fetchedAt := time.Now().UTC()

		savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
		for _, channelDetail := range channelDetails {
			ch, err := channel.NewChannel(fetchedAt, fetchedAt.AddDate(-1, 0, 0), channelDetail)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new channel(create playlist with access token)", slog.Any("error", err))
				continue
			}
			if _, err := channel.NewChannelRepository(sqlc.New(s.db)).Save(ctx, ch); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save channel(create playlist with access token)", slog.Any("error", err))
				continue
			}
			savedChannels[channelDetail.ID] = ch.ID
		}

		videoIDToInt64 := make(map[youtube_d.VideoID]int64)
		for _, videoDetail := range videoDetails {
			channelUUID, ok := savedChannels[videoDetail.ChannelID]
			if !ok {
				util.LoggerFromContext(ctx).InfoContext(ctx, "channel not found in saved channels(create playlist with access token)", slog.String("channelID", string(videoDetail.ChannelID)))
				continue
			}
			v, err := video.NewVideo(channelUUID, fetchedAt, videoDetail)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(create playlist with access token)", slog.Any("error", err))
				continue
			}
			q := sqlc.New(s.db)
			savedVideoID, err := video.NewVideoRepository(q).Save(ctx, v)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(create playlist with access token)", slog.Any("error", err))
				continue
			}
			if err := s.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
				util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(create playlist with access token)", slog.Any("error", err))
			}
			videoIDToInt64[v.Video.ID] = savedVideoID
		}

		for _, vid := range videoIDs {
			savedVideoID, ok := videoIDToInt64[vid]
			if !ok {
				util.LoggerFromContext(ctx).InfoContext(ctx, "video not found in saved videos(create playlist with access token)", slog.String("videoID", string(vid)))
				continue
			}
			allVideoIDs = append(allVideoIDs, savedVideoID)
		}

		if pageToken == "" {
			break
		}
		nextPageToken = pageToken
	}

	if err := NewPlaylistRepository(q).BulkInsertVideos(ctx, row, allVideoIDs); err != nil {
		return nil, err
	}

	if err := playlist.SetVideoCount(len(allVideoIDs)); err != nil {
		return nil, err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *Service) GetPlaylists(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetPlaylistsView, _ bool, err error) {
	defer util.Wrap(&err, "playlist.(*Service).GetPlaylists")

	playlists, err := s.playlistQS.FindPlaylists(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(playlists) > int(limit) {
		return playlists[:limit], true, nil
	}
	return playlists, false, nil
}

func (s *Service) GetChannelPlaylists(ctx context.Context, channelID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetChannelPlaylistsView, _ bool, err error) {
	defer util.Wrap(&err, "playlist.(*Service).GetChannelPlaylists")

	playlists, err := s.playlistQS.FindChannelPlaylists(ctx, channelID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(playlists) > int(limit) {
		return playlists[:limit], true, nil
	}
	return playlists, false, nil
}

func (s *Service) GetRecentPlaylists(ctx context.Context, userID uuid.UUID) (_ []GetChannelPlaylistsView, err error) {
	defer util.Wrap(&err, "playlist.(*Service).GetRecentPlaylists")

	return s.playlistQS.FindRecentPlaylists(ctx, userID)
}

func (s *Service) GetPlaylistDetail(ctx context.Context, userID, playlistID uuid.UUID) (_ GetPlaylistDetailView, err error) {
	defer util.Wrap(&err, "playlist.(*Service).GetPlaylistDetail")

	view, err := s.playlistQS.Find(ctx, userID, playlistID)
	if err != nil {
		return GetPlaylistDetailView{}, err
	}
	return view, nil
}

func (s *Service) DeletePlaylist(ctx context.Context, userID, playlistID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*Service).DeletePlaylist")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !playlist.IsModifiable() {
		return ErrPlaylistNotModifiable
	}

	if err := NewPlaylistRepository(q).Remove(ctx, userID, playlistID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) GetPlaylistItems(ctx context.Context, userID, playlistID uuid.UUID, videoCursor *uuid.UUID, limit int32) (_ []GetPlaylistItemView, _ bool, err error) {
	defer util.Wrap(&err, "playlist.(*Service).GetPlaylistItems")

	items, err := s.playlistQS.FindPlaylistItems(ctx, userID, playlistID, videoCursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(items) > int(limit) {
		return items[:limit], true, nil
	}
	return items, false, nil
}

func (s *Service) UpdatePlaylist(ctx context.Context, userID, playlistID uuid.UUID, newPlaylistTitle *string, newPlaylistDescription *string) (_ *Playlist, err error) {
	defer util.Wrap(&err, "playlist.(*Service).UpdatePlaylist")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}
	if !playlist.IsModifiable() {
		return nil, ErrPlaylistNotModifiable
	}

	if err := playlist.SetTitle(newPlaylistTitle); err != nil {
		return nil, err
	}
	if err := playlist.SetDescription(newPlaylistDescription); err != nil {
		return nil, err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *Service) InsertVideoIntoPlaylist(ctx context.Context, userID, playlistID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*Service).InsertVideoIntoPlaylist")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !playlist.IsModifiable() {
		return ErrPlaylistNotModifiable
	}
	if playlist.VideoCount >= 128 {
		return ErrTooManyPlaylistVideos
	}

	if err := NewPlaylistRepository(q).InsertVideo(ctx, userID, playlistID, videoID); err != nil {
		return err
	}

	playlist.IncrementVideoCount()

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) FetchAndInsertVideoIntoPlaylist(ctx context.Context, userID, playlistID uuid.UUID, externalVideoText string) (_ uuid.UUID, err error) {
	defer util.Wrap(&err, "playlist.(*Service).FetchAndInsertVideoIntoPlaylist")

	// YouTube動画IDを抽出
	ytVideoID, err := youtube_d.ExtractVideoID(externalVideoText)
	if err != nil {
		return uuid.Nil, err
	}

	vid, err := youtube_d.NewVideoID(ytVideoID)
	if err != nil {
		return uuid.Nil, err
	}

	// YouTube APIで動画の詳細を取得
	videoDetails, err := s.youtubeClient.FetchVideoDetail(ctx, []youtube_d.VideoID{vid})
	if err != nil {
		return uuid.Nil, err
	}
	vd, ok := videoDetails[vid]
	if !ok {
		return uuid.Nil, youtube_d.ErrInvalidVideoID
	}

	// YouTube APIでチャンネルの詳細を取得
	channelDetails, err := s.youtubeClient.FetchChannelDetail(ctx, []youtube_d.ChannelID{vd.ChannelID})
	if err != nil {
		return uuid.Nil, err
	}
	cd, ok := channelDetails[vd.ChannelID]
	if !ok {
		return uuid.Nil, youtube_d.ErrInvalidVideoID
	}

	fetchedAt := time.Now().UTC()
	q := sqlc.New(s.db)

	// チャンネルを挿入
	ch, err := channel.NewChannel(fetchedAt, fetchedAt.AddDate(-1, 0, 0), cd)
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := channel.NewChannelRepository(q).Save(ctx, ch); err != nil {
		return uuid.Nil, err
	}

	// 動画をupsert
	v, err := video.NewVideo(ch.ID, fetchedAt, vd)
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
		return uuid.Nil, err
	}
	if err := s.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
		util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(insert video by url)", slog.Any("error", err))
	}

	// プレイリストに追加
	if err := s.InsertVideoIntoPlaylist(ctx, userID, playlistID, v.ID); err != nil {
		return uuid.Nil, err
	}

	return v.ID, nil
}

func (s *Service) CopyPlaylist(ctx context.Context, userID uuid.UUID, sourcePlaylistID uuid.UUID, title, description string) (_ *Playlist, err error) {
	defer util.Wrap(&err, "playlist.(*Service).CopyPlaylist(userID=%s, sourcePlaylistID=%s)", userID, sourcePlaylistID)

	playlist, err := NewPlaylist(
		userID,
		title,
		description,
		"private",
		"normal",
	)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()

	q := sqlc.New(tx)
	repo := NewPlaylistRepository(q)

	count, err := q.CountUserPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 20 {
		return nil, ErrTooManyPlaylists
	}

	destInternalID, err := repo.Save(ctx, playlist)
	if err != nil {
		return nil, err
	}

	copiedCount, err := repo.CopyVideos(ctx, userID, sourcePlaylistID, destInternalID)
	if err != nil {
		return nil, err
	}

	if err := playlist.SetVideoCount(copiedCount); err != nil {
		return nil, err
	}

	if _, err := repo.Save(ctx, playlist); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *Service) RemoveVideoFromPlaylist(ctx context.Context, userID, playlistID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*Service).RemoveVideoFromPlaylist")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !playlist.IsModifiable() {
		return ErrPlaylistNotModifiable
	}

	if err := NewPlaylistRepository(q).RemoveVideo(ctx, userID, playlistID, videoID); err != nil {
		return err
	}

	if err := playlist.DecrementVideoCount(); err != nil {
		return err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) MarkAsWatchLater(ctx context.Context, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*Service).MarkAsWatchLater")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// watch laterのプレイリストをロッキングリード
	playlist, err := NewPlaylistRepository(q).FindWatchLaterForUpdate(ctx, userID)
	if err != nil {
		return err
	}

	// 動画をinsert
	if err := NewPlaylistRepository(q).InsertWatchLater(ctx, userID, playlist.ID, videoID); err != nil {
		return err
	}

	// 動画をインクリメントして保存
	playlist.IncrementVideoCount()

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) UnmarkAsWatchLater(ctx context.Context, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "playlist.(*Service).UnmarkAsWatchLater")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// watch laterのプレイリストをロッキングリード
	playlist, err := NewPlaylistRepository(q).FindWatchLaterForUpdate(ctx, userID)
	if err != nil {
		return err
	}

	// 動画をremove
	if err := NewPlaylistRepository(q).RemoveVideo(ctx, userID, playlist.ID, videoID); err != nil {
		return err
	}

	// 動画をデクリメントて保存
	if err := playlist.DecrementVideoCount(); err != nil {
		return err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, playlist); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
