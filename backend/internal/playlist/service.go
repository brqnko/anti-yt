package playlist

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidPlaylistID = errors.New("invalid playlist id or unsupported format")

// extractPlaylistID はYouTubeのプレイリストURLまたは生のプレイリストIDからプレイリストIDを抽出する
// NOTE: UU, PL, OL以外はAPIで取得できないので無視する.
// UUはチャンネルのアップロード動画
// PLはプレイリスt
// OLは公式のプレイリスト
func extractPlaylistID(playlistText string) (string, error) {
	if strings.HasPrefix(playlistText, "PL") || strings.HasPrefix(playlistText, "UU") || strings.HasPrefix(playlistText, "OL") {
		return playlistText, nil
	}

	if !strings.HasPrefix(playlistText, "http://") && !strings.HasPrefix(playlistText, "https://") {
		playlistText = "https://" + playlistText
	}
	u, err := url.Parse(playlistText)
	if err != nil {
		return "", ErrInvalidPlaylistID
	}

	if u.Host != "youtube.com" && !strings.HasSuffix(u.Host, ".youtube.com") && u.Host != "youtu.be" {
		return "", ErrInvalidPlaylistID
	}

	listID := u.Query().Get("list")
	if listID == "" {
		return "", ErrInvalidPlaylistID
	}

	return listID, nil
}

type Service struct {
	db        *pgxpool.Pool
	ytService youtube_d.YouTubeAPIService
}

func NewService(db *pgxpool.Pool, ytService youtube_d.YouTubeAPIService) (*Service, error) {
	return &Service{
		db:        db,
		ytService: ytService,
	}, nil
}

func (s *Service) CreatePlaylist(ctx context.Context, title, description, visibilityStr, playlistTypeStr string, basePlaylistUrl *string) (*Playlist, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pl, err := NewPlaylist(
		uuid.Nil,
		title,
		description,
		visibilityStr,
		playlistTypeStr,
		0,
		time.Time{},
		nil,
		[]*video.Video{},
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

	// NOTE: repeatable readなので、insertした内容はcommitするまで他のトランザクションから見れない
	// なので、ロックなどをつける意味はない
	row, err := q.CreatePlaylist(ctx, sqlc.CreatePlaylistParams{
		UserPublicID:        userID,
		PlaylistTitle:       string(*pl.Title),
		PlaylistDescription: string(*pl.Description),
		VisibilityCode:      int(pl.VisibilityCode),
		PlaylistCode:        int(pl.PlaylistCode),
	})
	if err != nil {
		return nil, fmt.Errorf("createPlaylist: %w", err)
	}

	pl.SetGeneratedFields(row.PublicID, row.CreatedAt)

	if basePlaylistUrl == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return pl, nil
	}

	// YouTubeからプレイリストをimportする
	playlistID, err := extractPlaylistID(*basePlaylistUrl)
	if err != nil {
		return nil, err
	}

	// insert時に使うcopyfromはon conflictに対応していないので、コード側で重複がないことを検証する
	var allVideoIDs []int64
	seenVideoIDs := make(map[int64]struct{})
	var nextPageToken string
	for {
		// プレイリストから動画IDのリストを取得
		videoIDs, pageToken, err := s.ytService.FetchPlaylistVideoIDs(ctx, playlistID, nextPageToken)
		if err != nil {
			return nil, fmt.Errorf("fetchPlaylistVideoIds: %w", err)
		}

		// 動画IDのリストから詳細を取得
		videoDetails, err := s.ytService.FetchVideoDetail(ctx, videoIDs)
		if err != nil {
			return nil, fmt.Errorf("fetchVideoDetail: %w", err)
		}

		// 各動画のチャンネル詳細を取得（重複排除）
		channelIDSet := make(map[string]struct{})
		for _, vd := range videoDetails {
			channelIDSet[vd.ChannelID] = struct{}{}
		}
		channelIDs := make([]string, len(channelIDSet))
		var j int
		for id := range channelIDSet {
			channelIDs[j] = id
			j++
		}
		channelDetails, err := s.ytService.FetchChannelDetail(ctx, channelIDs)
		if err != nil {
			return nil, fmt.Errorf("fetchChannelDetail: %w", err)
		}

		savedChannels := make(map[string]int64)

		// NOTE: チャンネルを先に保存するので、動画があってチャンネルがないようなことは発生しない
		// そのため、トランザクションは使用しない.
		// どのチャンネルがどんな動画を投稿したかのデータの、プレイリストの保存の成功に関わらず蓄積されてく
		for _, channelDetail := range channelDetails {
			if err := sqlc.New(s.db).ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
				ExternalCustomID: channelDetail.CustomID,
				ExternalID:       channelDetail.ID,
			}); err != nil {
				return nil, fmt.Errorf("clearStaleChannelCustomID: %w", err)
			}
			saveChannel, err := sqlc.New(s.db).SaveChannel(ctx, sqlc.SaveChannelParams{
				ExternalID:                channelDetail.ID,
				ExternalDisplayName:       channelDetail.DisplayName,
				ExternalCustomID:          channelDetail.CustomID,
				ExternalIconUrl:           channelDetail.IconURL,
				ExternalDescription:       channelDetail.Description,
				ExternalSubscribersCount:  int64(channelDetail.SubscribersCount),
				ExternalCreatedAt:         channelDetail.CreatedAt,
				ExternalUploadsPlaylistID: channelDetail.UploadsPlaylistID,
			})
			if err != nil { // NOTE: on conflict do update returningはコンフリクト時はその行を返してれる
				return nil, err
			}
			savedChannels[channelDetail.ID] = saveChannel.MChannelID
		}

		fetchedAt := time.Now().UTC()
		for _, vid := range videoIDs {
			videoDetail, ok := videoDetails[vid]
			if !ok {
				continue
			}
			saveVideo, err := sqlc.New(s.db).SaveVideo(ctx, sqlc.SaveVideoParams{
				MChannelID:            savedChannels[videoDetail.ChannelID],
				ExternalID:            videoDetail.ID,
				ExternalTitle:         videoDetail.Title,
				ExternalDescription:   videoDetail.Description,
				FetchedAt:             fetchedAt,
				ExternalCreatedAt:     videoDetail.CreatedAt,
				ExternalThumbnailUrl:  videoDetail.ThumbnailURL,
				ExternalLengthSeconds: videoDetail.LengthSeconds,
			})
			if err != nil {
				return nil, fmt.Errorf("saveVideo: %w", err)
			}

			if _, exists := seenVideoIDs[saveVideo]; !exists {
				seenVideoIDs[saveVideo] = struct{}{}
				allVideoIDs = append(allVideoIDs, saveVideo)
			}
		}

		if pageToken == "" {
			break
		}
		nextPageToken = pageToken
	}

	bulkParams := make([]sqlc.BulkInsertIntoPlaylistParams, len(allVideoIDs))
	for i, videoID := range allVideoIDs {
		bulkParams[i] = sqlc.BulkInsertIntoPlaylistParams{
			MPlaylistID:      row.MPlaylistID,
			MVideoID:         videoID,
			PlaylistPosition: int64(i) * 1048576, // 2^20
		}
	}
	if _, err := q.BulkInsertIntoPlaylist(ctx, bulkParams); err != nil {
		return nil, fmt.Errorf("bulkInsertIntoPlaylist: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return pl, nil
}

func (s *Service) GetPlaylists(ctx context.Context, cursor *uuid.UUID, limit int) ([]*Playlist, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(s.db)
	rows, err := q.GetUserPlaylists(ctx, sqlc.GetUserPlaylistsParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("getUserPlaylists: %w", err)
	}

	playlists := make([]*Playlist, len(rows))
	for i, row := range rows {
		var topThumbnail *string
		if row.TopThumbnail != "" {
			topThumbnail = &row.TopThumbnail
		}

		pl, err := NewPlaylist(
			row.PublicID,
			row.PlaylistTitle,
			row.PlaylistDescription,
			VisibilityCode(row.VisibilityCode).String(),
			PlaylistCode(row.PlaylistCode).String(),
			int(row.VideoCount),
			row.CreatedAt,
			topThumbnail,
			[]*video.Video{},
		)
		if err != nil {
			return nil, fmt.Errorf("newPlaylist: %w", err)
		}
		playlists[i] = pl
	}

	return playlists, nil
}

func (s *Service) GetPlaylistInfo(ctx context.Context, playlistID uuid.UUID) (*Playlist, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(s.db)
	row, err := q.GetPlaylist(ctx, sqlc.GetPlaylistParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPlaylistNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getPlaylistInfo: %w", err)
	}

	var topThumbnail *string
	if row.TopThumbnail != "" {
		topThumbnail = &row.TopThumbnail
	}

	return NewPlaylist(
		row.PublicID,
		row.PlaylistTitle,
		row.PlaylistDescription,
		VisibilityCode(row.VisibilityCode).String(),
		PlaylistCode(row.PlaylistCode).String(),
		int(row.VideoCount),
		row.CreatedAt,
		topThumbnail,
		[]*video.Video{},
	)
}

func (s *Service) DeletePlaylist(ctx context.Context, playlistID uuid.UUID) error {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	q := sqlc.New(s.db)
	_, err = q.DeletePlaylist(ctx, sqlc.DeletePlaylistParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPlaylistNotFound
	}
	if err != nil {
		return fmt.Errorf("deletePlaylist: %w", err)
	}

	return nil
}

func (s *Service) GetPlaylistItems(ctx context.Context, playlistID uuid.UUID, videoCursor *uuid.UUID, limit int) ([]*video.Video, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(s.db)
	rows, err := q.GetPlaylistVideos(ctx, sqlc.GetPlaylistVideosParams{
		UserID:     userID,
		PlaylistID: playlistID,
		Cursor:     videoCursor,
		QueryLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("getPlaylistVideos: %w", err)
	}

	videos := make([]*video.Video, len(rows))
	for i, row := range rows {
		videos[i] = video.NewVideo(
			row.PublicID,
			row.ChannelID,
			row.ExternalThumbnailUrl,
			row.ExternalChannelIconUrl,
			row.ExternalTitle,
			row.ExternalChannelDisplayname,
			row.ExternalCreatedAt,
			row.ExternalLengthSeconds,
			row.LastWatchSeconds,
		)
	}

	return videos, nil
}

func (s *Service) UpdatePlaylist(ctx context.Context, playlistID uuid.UUID, newPlaylistTitle *string, newPlaylistDescription *string) (*Playlist, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if newPlaylistTitle != nil {
		if _, err := NewPlaylistTitle(*newPlaylistTitle); err != nil {
			return nil, err
		}
	}
	if newPlaylistDescription != nil {
		if _, err := NewPlaylistDescription(*newPlaylistDescription); err != nil {
			return nil, err
		}
	}

	q := sqlc.New(s.db)
	_, err = q.UpdatePlaylist(ctx, sqlc.UpdatePlaylistParams{
		UserID:                 userID,
		PlaylistID:             playlistID,
		NewPlaylistTitle:       newPlaylistTitle,
		NewPlaylistDescription: newPlaylistDescription,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPlaylistNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updatePlaylist: %w", err)
	}

	row, err := q.GetPlaylist(ctx, sqlc.GetPlaylistParams{
		UserID:     userID,
		PlaylistID: playlistID,
	})
	if err != nil {
		return nil, fmt.Errorf("getPlaylist: %w", err)
	}

	var topThumbnail *string
	if row.TopThumbnail != "" {
		topThumbnail = &row.TopThumbnail
	}

	return NewPlaylist(
		row.PublicID,
		row.PlaylistTitle,
		row.PlaylistDescription,
		VisibilityCode(row.VisibilityCode).String(),
		PlaylistCode(row.PlaylistCode).String(),
		int(row.VideoCount),
		row.CreatedAt,
		topThumbnail,
		[]*video.Video{},
	)
}

func (s *Service) InsertVideoIntoPlaylist(ctx context.Context, playlistID uuid.UUID, videoID uuid.UUID) (time.Time, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return time.Time{}, err
	}

	q := sqlc.New(s.db)
	row, err := q.InsertIntoPlaylist(ctx, sqlc.InsertIntoPlaylistParams{
		UserID:     userID,
		PlaylistID: playlistID,
		VideoID:    videoID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrPlaylistNotFound
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("insertIntoPlaylist: %w", err)
	}

	return row.UpdatedAt, nil
}

func (s *Service) RemoveVideoFromPlaylist(ctx context.Context, playlistID uuid.UUID, videoID uuid.UUID) error {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	q := sqlc.New(s.db)
	_, err = q.RemoveVideoFromPlaylist(ctx, sqlc.RemoveVideoFromPlaylistParams{
		UserID:     userID,
		PlaylistID: playlistID,
		VideoID:    videoID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrVideoNotInPlaylist
	}
	if err != nil {
		return fmt.Errorf("removeVideoFromPlaylist: %w", err)
	}

	return nil
}
