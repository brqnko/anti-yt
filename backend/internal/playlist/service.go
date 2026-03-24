package playlist

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

var ErrInvalidPlaylistID = util.NewDomainError("playlist.invalid_playlist_id", "invalid playlist id or unsupported format")

type Service struct {
	db         *pgxpool.Pool
	ytService  youtube_d.YouTubeAPIService
	playlistQS PlaylistQueryService
}

func NewService(db *pgxpool.Pool, ytService youtube_d.YouTubeAPIService) (*Service, error) {
	return &Service{
		db:         db,
		ytService:  ytService,
		playlistQS: NewPlaylistQueryService(db),
	}, nil
}

func (s *Service) CreatePlaylist(ctx context.Context, userID uuid.UUID, title, description, visibilityStr, playlistTypeStr string, basePlaylistUrl *string) (_ *Playlist, err error) {
	defer util.Wrap(&err, "Service.CreatePlaylist(userID=%s)", userID)
	playlist, err := NewPlaylist(
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
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	// NOTE: insertした内容はcommitするまで他のトランザクションから見れない(repeatable read)ので、ロックなどをつける意味はない
	row, err := NewPlaylistRepository(q).Save(ctx, userID, playlist)
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
	playlistID, err := extractPlaylistID(*basePlaylistUrl)
	if err != nil {
		return nil, err
	}

	var allVideoIDs []int64 // 順序を保持するため
	var nextPageToken string
	for {
		// プレイリストから動画IDのリストを取得
		videoIDs, pageToken, err := s.ytService.FetchPlaylistVideoIDs(ctx, playlistID, nextPageToken)
		if err != nil {
			return nil, err
		}

		// 動画IDのリストから詳細を取得
		videoDetails, err := s.ytService.FetchVideoDetail(ctx, videoIDs)
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
		channelDetails, err := s.ytService.FetchChannelDetail(ctx, channelIDs)
		if err != nil {
			return nil, err
		}

		fetchedAt := time.Now().UTC()

		// NOTE: チャンネルを先に保存するので、動画があってチャンネルがないようなことは発生しないため、トランザクションは使用しない
		// チャンネルの詳細情報を先に保存する
		savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
		for _, channelDetail := range channelDetails {
			ch, err := channel.NewChannel(fetchedAt, fetchedAt, channelDetail)
			if err != nil {
				return nil, err
			}

			if _, err := channel.NewChannelRepository(sqlc.New(s.db)).Save(ctx, ch); err != nil {
				return nil, err
			}
			savedChannels[channelDetail.ID] = ch.ID
		}

		// 動画の詳細情報を保存する
		// 動画の保存自体に順番は関係ない
		videoIDToint64 := make(map[youtube_d.VideoID]int64)
		for _, videoDetail := range videoDetails {
			v, err := video.NewVideo(savedChannels[videoDetail.ChannelID], fetchedAt, videoDetail)
			if err != nil {
				slog.Info("failed to newVideo(createPlaylist)", "error", err)
				continue
			}

			savedVideoID, err := video.NewVideoRepository(sqlc.New(s.db)).Save(ctx, v)
			if err != nil {
				slog.Info("failed to saveVideo(createPlaylist)", "error", err)
				continue
			}
			videoIDToint64[v.Video.ID] = savedVideoID
		}

		for _, vid := range videoIDs {
			savedVideoID, ok := videoIDToint64[vid]
			if !ok {
				slog.Info("video not found in savedVideos(createPlaylist)", "videoID", vid)
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

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *Service) GetPlaylists(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (playlists []GetPlaylistsView, hasNext bool, err error) {
	playlists, err = s.playlistQS.FindPlaylists(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(playlists) > int(limit) {
		return playlists[:limit], true, nil
	}
	return playlists, false, nil
}

func (s *Service) GetPlaylistDetail(ctx context.Context, userID, playlistID uuid.UUID) (GetPlaylistDetailView, error) {
	view, err := s.playlistQS.Find(ctx, userID, playlistID)
	if err != nil {
		return GetPlaylistDetailView{}, err
	}
	return view, nil
}

func (s *Service) DeletePlaylist(ctx context.Context, userID, playlistID uuid.UUID) error {
	return NewPlaylistRepository(sqlc.New(s.db)).Remove(ctx, userID, playlistID)
}

func (s *Service) GetPlaylistItems(ctx context.Context, userID, playlistID uuid.UUID, videoCursor *uuid.UUID, limit int32) (items []GetPlaylistItemView, hasNext bool, err error) {
	items, err = s.playlistQS.FindPlaylistItems(ctx, userID, playlistID, videoCursor, limit+1)
	if err != nil {
		return nil, false, err
	}

	if len(items) > int(limit) {
		return items[:limit], true, nil
	}
	return items, false, nil
}

func (s *Service) UpdatePlaylist(ctx context.Context, userID, playlistID uuid.UUID, newPlaylistTitle *string, newPlaylistDescription *string) (*Playlist, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}

	if err := playlist.SetTitle(newPlaylistTitle); err != nil {
		return nil, err
	}
	if err := playlist.SetDescription(newPlaylistDescription); err != nil {
		return nil, err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, userID, playlist); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (s *Service) InsertVideoIntoPlaylist(ctx context.Context, userID, playlistID, videoID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return err
	}

	if err := NewPlaylistRepository(q).InsertVideo(ctx, userID, playlistID, videoID); err != nil {
		return err
	}

	playlist.IncrementVideoCount()

	if _, err := NewPlaylistRepository(q).Save(ctx, userID, playlist); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) RemoveVideoFromPlaylist(ctx context.Context, userID, playlistID, videoID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()

	q := sqlc.New(tx)
	playlist, err := NewPlaylistRepository(q).FindForUpdate(ctx, userID, playlistID)
	if err != nil {
		return err
	}

	if err := NewPlaylistRepository(q).RemoveVideo(ctx, userID, playlistID, videoID); err != nil {
		return err
	}

	if err := playlist.DecrementVideoCount(); err != nil {
		return err
	}

	if _, err := NewPlaylistRepository(q).Save(ctx, userID, playlist); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
