package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exhaustQuotaJob struct {
	db             *pgxpool.Pool
	ytService      youtube_d.Service
	discordService discord_d.Service
	mx             *sync.Mutex
}

func (j *exhaustQuotaJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*exhaustQuotaJob).run")

	tx, err := j.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback in exhaustQuotaJob.run", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// ad lock
	if err := database_d.TryAdLock(ctx, q, []byte("exhaustQuotaJob")); err != nil {
		return err
	}

	// TODO
	channels, err := channel.NewChannelRepository(q).FindBulkFetchedAfter(ctx, time.Now().UTC().Add(24*30*time.Hour))
	if err != nil {
		return err
	}
	for _, c := range channels {
		if err := database_d.TryAdLock(ctx, q, c.ID[:]); err != nil {
			util.LoggerFromContext(ctx).WarnContext(ctx, "failed to acquire ad lock for channel", slog.String("channel_id", c.ID.String()), slog.Any("error", err))
			continue
		}

		// 動画を全て取得して保存する
		fetchedAt := time.Now().UTC()
		pageToken := ""
		for {
			videoIDs, nextPageToken, err := j.ytService.FetchPlaylistVideoIDs(ctx, string(c.Channel.UploadsPlaylistID), pageToken)
			if err != nil {
				return err
			}

			if len(videoIDs) > 0 {
				videoDetails, err := j.ytService.FetchVideoDetail(ctx, videoIDs)
				if err != nil {
					util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to fetch video detail", slog.Any("error", err))
				} else {
					for _, vd := range videoDetails {
						v, err := video.NewVideo(c.ID, fetchedAt, vd)
						if err != nil {
							util.LoggerFromContext(ctx).InfoContext(ctx, "failed to newVideo", slog.Any("error", err))
							continue
						}
						if _, err := video.NewVideoRepository(q).Save(ctx, v); err != nil {
							util.LoggerFromContext(ctx).InfoContext(ctx, "failed to save video", slog.Any("error", err))
							continue
						}
					}
				}
			}

			if nextPageToken == "" {
				break
			}
			pageToken = nextPageToken
		}

		// プレイリストを全て取得

		var ytPlaylists []youtube_d.Playlist
		pageToken = ""
		for {
			playlists, nextPageToken, err := j.ytService.FetchChannelPlaylists(ctx, c.Channel.ID, pageToken)
			if err != nil {
				return err
			}
			ytPlaylists = append(ytPlaylists, playlists...)

			if nextPageToken == "" {
				break
			}
			pageToken = nextPageToken
		}

		for _, ytPlaylist := range ytPlaylists {
			util.LoggerFromContext(ctx).InfoContext(ctx, "importing playlist", slog.String("playlistID", ytPlaylist.ID), slog.String("title", ytPlaylist.Title))

			// プレイリストの動画ID一覧を全ページ取得
			var allVideoIDs []youtube_d.VideoID
			videoPageToken := ""
			for {
				videoIDs, nextVideoPageToken, err := j.ytService.FetchPlaylistVideoIDs(ctx, ytPlaylist.ID, videoPageToken)
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
					videoDetails, err := j.ytService.FetchVideoDetail(ctx, toRequestVideos)
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
				videoDetails, err := j.ytService.FetchVideoDetail(ctx, toRequestVideos)
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
					channelDetails, err := j.ytService.FetchChannelDetail(ctx, toRequestChannels)
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
				channelDetails, err := j.ytService.FetchChannelDetail(ctx, toRequestChannels)
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
				uuid.Nil,
				ytPlaylist.Title,
				ytPlaylist.Description,
				"public",
				"external_auto",
				playlist.WithPlaylistRegisteredAt(ytPlaylist.CreatedAt),
				playlist.WithPlaylistChannelID(c.ID),
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
		}

		// チャンネルをbulk fetched済みとしてマーク
		c.MarkAsBulkFetched()
		c.MarkAsRSSFetched()

		if _, err := channel.NewChannelRepository(q).Save(ctx, c); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to save channel", slog.Any("error", err))
			continue
		}

		// release ad lock
		if err := database_d.ReleaseAdLock(ctx, q, c.ID[:]); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to release ad lock for channel", slog.String("channel_id", c.ID.String()), slog.Any("error", err))
			continue
		}

		// discord webhookに送信
		if err := j.discordService.SendWebhookMessage(ctx, fmt.Sprintf("チャンネル %s のバルクフェッチが完了しました", c.ID.String())); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to send discord webhook", slog.String("channel_id", c.ID.String()), slog.Any("error", err))
		}
	}

	if err := tx.Commit(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return err
	}

	return nil
}

func (j *exhaustQuotaJob) Run() {
	// クオータリセットはPT midnight。cronは夏冬両方で登録されるので、
	// リセットまで10分以内でなければスキップする。
	loc, _ := time.LoadLocation("America/Los_Angeles")
	now := time.Now().In(loc)
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	if nextMidnight.Sub(now) > 15*time.Minute {
		slog.Info("skipping exhaust quota job: not close enough to quota reset")
		return
	}

	j.mx.Lock()
	defer j.mx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run exhaust quota job", slog.Any("error", err))
	}
}

func NewExhaustQuotaJob(db *pgxpool.Pool, ytService youtube_d.Service, discordService discord_d.Service) scheduler.Job {
	return &exhaustQuotaJob{
		db:             db,
		ytService:      ytService,
		discordService: discordService,
		mx:             &sync.Mutex{},
	}
}
