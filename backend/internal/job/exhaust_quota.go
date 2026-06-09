package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exhaustQuotaJob struct {
	db            *pgxpool.Pool
	youtubeClient youtube_d.Client
	feedRepo      database_d.FeedRepository
	discordClient discord_d.Client
	mx            *sync.Mutex
}

func (j *exhaustQuotaJob) run(ctx context.Context) (processed int, err error) {
	defer util.Wrap(&err, "job.(*exhaustQuotaJob).run")

	if wErr := j.discordClient.SendWebhookMessage(ctx, "**[Exhaust Quota]** Started"); wErr != nil {
		slog.Error("failed to send discord webhook(exhaust quota job start)", slog.Any("error", wErr))
	}

	q := sqlc.New(j.db)

	// セッションレベルのadvisory lockを取得
	if err := database_d.TryAdLockSession(ctx, q, []byte("exhaustQuotaJob")); err != nil {
		return 0, err
	}
	defer func() {
		if err := database_d.ReleaseAdLock(ctx, q, []byte("exhaustQuotaJob")); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to release ad lock(exhaust quota job)", slog.Any("error", err))
		}
	}()

	channels, err := channel.NewChannelRepository(q).ListForBulkFetch(ctx)
	if err != nil {
		return 0, err
	}
	quotaExhausted := false
channelLoop:
	for _, c := range channels {
		if err := database_d.TryAdLockSession(ctx, q, c.ID[:]); err != nil {
			util.LoggerFromContext(ctx).WarnContext(ctx, "failed to acquire ad lock for channel", slog.String("channel_id", c.ID.String()), slog.Any("error", err))
			continue
		}

		// 動画を全て取得して保存する
		// uploads playlistはYouTubeから新しい順で返ってくるので、前回のbulk_fetched_atより
		// 古い動画に到達したらそれ以降は前回までで取得済みとみなしてbreakする
		prevBulkFetchedAt := c.BulkFetchedAt
		fetchedAt := time.Now().UTC()
		pageToken := ""
		reachedPrevBulkFetch := false
		for {
			videoIDs, nextPageToken, err := j.youtubeClient.FetchPlaylistVideoIDs(ctx, string(c.Channel.UploadsPlaylistID), pageToken)
			if err != nil {
				if relErr := database_d.ReleaseAdLock(ctx, q, c.ID[:]); relErr != nil {
					util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to release ad lock for channel", slog.String("channel_id", c.ID.String()), slog.Any("error", relErr))
				}
				if errors.Is(err, youtube_d.ErrQuotaExceeded) {
					quotaExhausted = true
					break channelLoop
				}
				util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to fetch uploads playlist video IDs(exhaust quota job)", slog.String("channel_id", c.ID.String()), slog.String("uploads_playlist_id", string(c.Channel.UploadsPlaylistID)), slog.Any("error", err))
				continue channelLoop
			}

			if len(videoIDs) > 0 {
				videoDetails, err := j.youtubeClient.FetchVideoDetail(ctx, videoIDs)
				if err != nil {
					util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to fetch video detail", slog.Any("error", err))
					continue
				}

				for _, vid := range videoIDs {
					vd, ok := videoDetails[vid]
					if !ok {
						continue
					}
					if vd.CreatedAt.Before(prevBulkFetchedAt) {
						reachedPrevBulkFetch = true
						break
					}
					v, err := video.NewVideo(c.ID, fetchedAt, vd)
					if err != nil {
						util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(exhaust quota job)", slog.Any("error", err))
						continue
					}
					q := sqlc.New(j.db)
					if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
						util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(exhaust quota job)", slog.Any("error", err))
						continue
					}
					if err := j.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
						util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(exhaust quota job)", slog.Any("error", err))
					}
				}
			}

			if reachedPrevBulkFetch || nextPageToken == "" {
				break
			}
			pageToken = nextPageToken
		}

		// プレイリストを全て取得
		var ytPlaylists []youtube_d.Playlist
		pageToken = ""
		fetchPlaylistsFailed := false
		for {
			playlists, nextPageToken, err := j.youtubeClient.FetchChannelPlaylists(ctx, c.Channel.ID, pageToken)
			if err != nil {
				if errors.Is(err, youtube_d.ErrQuotaExceeded) {
					quotaExhausted = true
					break channelLoop
				}
				util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to fetch channel playlists(exhaust quota job)", slog.String("channel_id", c.ID.String()), slog.String("yt_channel_id", string(c.Channel.ID)), slog.Any("error", err))
				fetchPlaylistsFailed = true
				break
			}
			ytPlaylists = append(ytPlaylists, playlists...)

			if nextPageToken == "" {
				break
			}
			pageToken = nextPageToken
		}

		if fetchPlaylistsFailed {
			if relErr := database_d.ReleaseAdLock(ctx, q, c.ID[:]); relErr != nil {
				util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to release ad lock for channel", slog.String("channel_id", c.ID.String()), slog.Any("error", relErr))
			}
			continue channelLoop
		}

		playlistQS := playlist.NewPlaylistQueryService(j.db)
		for _, ytPlaylist := range ytPlaylists {
			// 既に取り込み済みのプレイリストはquota節約のためスキップする
			exists, err := playlistQS.ExistsByExternalID(ctx, ytPlaylist.ID)
			if err != nil {
				util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to check playlist existence(exhaust quota job)", slog.String("playlistID", ytPlaylist.ID), slog.Any("error", err))
				continue
			}
			if exists {
				util.LoggerFromContext(ctx).InfoContext(ctx, "skipping already fetched playlist", slog.String("playlistID", ytPlaylist.ID), slog.String("title", ytPlaylist.Title))
				continue
			}

			util.LoggerFromContext(ctx).InfoContext(ctx, "importing playlist", slog.String("playlistID", ytPlaylist.ID), slog.String("title", ytPlaylist.Title))

			// プレイリストの動画ID一覧を全ページ取得
			var allVideoIDs []youtube_d.VideoID
			videoPageToken := ""
			for {
				videoIDs, nextVideoPageToken, err := j.youtubeClient.FetchPlaylistVideoIDs(ctx, ytPlaylist.ID, videoPageToken)
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
					videoDetails, err := j.youtubeClient.FetchVideoDetail(ctx, toRequestVideos)
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
				videoDetails, err := j.youtubeClient.FetchVideoDetail(ctx, toRequestVideos)
				if err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch video details", slog.Any("error", err))
					continue
				}
				for id, vd := range videoDetails {
					videoDetailMap[id] = vd
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
					channelDetails, err := j.youtubeClient.FetchChannelDetail(ctx, toRequestChannels)
					toRequestChannels = nil
					if err != nil {
						util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch channel details", slog.Any("error", err))
						continue
					}

					for id, cd := range channelDetails {
						channelDetailMap[id] = cd
					}
				}
			}
			if len(toRequestChannels) > 0 {
				channelDetails, err := j.youtubeClient.FetchChannelDetail(ctx, toRequestChannels)
				if err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to fetch channel details", slog.Any("error", err))
					continue
				}

				for id, cd := range channelDetails {
					channelDetailMap[id] = cd
				}
			}

			fetchedAt := time.Now().UTC()

			// チャンネルを保存
			savedChannels := make(map[youtube_d.ChannelID]uuid.UUID)
			for _, cd := range channelDetailMap {
				ch, err := channel.NewChannel(fetchedAt, fetchedAt.AddDate(-1, 0, 0), cd)
				if err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new channel(exhaust quota job)", slog.Any("error", err))
					continue
				}
				if _, err := channel.NewChannelRepository(sqlc.New(j.db)).Save(ctx, ch); err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save channel(exhaust quota job)", slog.Any("error", err))
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
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to new video(exhaust quota job)", slog.Any("error", err))
					continue
				}
				q := sqlc.New(j.db)
				savedVideoID, err := video.NewVideoRepository(q).Save(ctx, v)
				if err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video(exhaust quota job)", slog.Any("error", err))
					continue
				}
				if err := j.feedRepo.FanOut(ctx, v.ChannelID, v.ID); err != nil {
					util.LoggerFromContext(ctx).WarnContext(ctx, "failed to fan-out video(exhaust quota job)", slog.Any("error", err))
				}
				savedVideos[v.Video.ID] = savedVideoID
			}

			// プレイリストを作成して動画を挿入
			description := ytPlaylist.Description
			if utf8.RuneCountInString(description) > 255 {
				description = string([]rune(description)[:255])
			}
			pl, err := playlist.NewPlaylist(
				uuid.Nil,
				ytPlaylist.Title,
				description,
				"public",
				"external_auto",
				playlist.WithPlaylistRegisteredAt(ytPlaylist.CreatedAt),
				playlist.WithPlaylistChannelID(c.ID),
				playlist.WithPlaylistExternalID(ytPlaylist.ID),
			)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to create playlist domain(exhaust quota job)", slog.Any("error", err))
				continue
			}

			playlistRow, err := playlist.NewPlaylistRepository(sqlc.New(j.db)).SaveSystem(ctx, pl)
			if err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save playlist(exhaust quota job)", slog.Any("error", err))
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
				if err := playlist.NewPlaylistRepository(sqlc.New(j.db)).BulkInsertVideos(ctx, playlistRow, videoInternalIDs); err != nil {
					util.LoggerFromContext(ctx).InfoContext(ctx, "failed to bulk insert videos(exhaust quota job)", slog.Any("error", err))
					continue
				}
			}

			if err := pl.SetVideoCount(len(videoInternalIDs)); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to set video count(exhaust quota job)", slog.Any("error", err))
				continue
			}

			if _, err := playlist.NewPlaylistRepository(sqlc.New(j.db)).SaveSystem(ctx, pl); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save playlist with video count(exhaust quota job)", slog.Any("error", err))
				continue
			}
		}

		// チャンネルをbulk fetched済みとしてマーク
		c.MarkAsBulkFetched()
		c.MarkAsRSSFetched()

		if _, err := channel.NewChannelRepository(sqlc.New(j.db)).Save(ctx, c); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to save channel(exhaust quota job)", slog.Any("error", err))
			continue
		}

		// チャンネルのad lockを解放
		if err := database_d.ReleaseAdLock(ctx, q, c.ID[:]); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to release ad lock for channel", slog.String("channel_id", c.ID.String()), slog.Any("error", err))
		}

		processed++
	}

	if quotaExhausted {
		return processed, youtube_d.ErrQuotaExceeded
	}
	return processed, nil
}

func (j *exhaustQuotaJob) Run() {
	// クオータリセットはPT midnight。cronは夏冬両方で登録されるので、
	// リセットまで1時間以内でなければスキップする。
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		slog.Error("failed to load location for exhaust quota job", slog.Any("error", err))
		return
	}
	now := time.Now().In(loc)
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	if nextMidnight.Sub(now) > 1*time.Hour {
		slog.Info("skipping exhaust quota job: not close enough to quota reset")
		return
	}

	j.mx.Lock()
	defer j.mx.Unlock()

	ctx, cancel := context.WithDeadline(context.Background(), nextMidnight)
	defer cancel()

	processed, err := j.run(ctx)

	notifyCtx, notifyCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer notifyCancel()

	var msg string
	switch {
	case errors.Is(err, youtube_d.ErrQuotaExceeded):
		msg = fmt.Sprintf("**[Exhaust Quota]** Quota exhausted\nProcessed: **%d** channels", processed)
	case err != nil && ctx.Err() == nil:
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run exhaust quota job", slog.Any("error", err))
		msg = fmt.Sprintf("[Error] exhaust quota job: %v\nProcessed: **%d** channels", err, processed)
	case ctx.Err() != nil:
		msg = fmt.Sprintf("**[Exhaust Quota]** Deadline reached (quota reset)\nProcessed: **%d** channels", processed)
	default:
		msg = fmt.Sprintf("**[Exhaust Quota]** Completed\nProcessed: **%d** channels", processed)
	}

	if wErr := j.discordClient.SendWebhookMessage(notifyCtx, msg); wErr != nil {
		slog.Error("failed to send discord webhook(exhaust quota job)", slog.Any("error", wErr))
	}
}

func NewExhaustQuotaJob(db *pgxpool.Pool, youtubeClient youtube_d.Client, feedRepo database_d.FeedRepository, discordClient discord_d.Client) scheduler.Job {
	return &exhaustQuotaJob{
		db:            db,
		youtubeClient: youtubeClient,
		feedRepo:      feedRepo,
		discordClient: discordClient,
		mx:            new(sync.Mutex{}),
	}
}
