package database_d

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type FeedRepository interface {
	Push(ctx context.Context, userIDs []uuid.UUID, videoID uuid.UUID) (err error)
	PushOne(ctx context.Context, userID uuid.UUID, videoID uuid.UUID) (err error)
	PushMany(ctx context.Context, userID uuid.UUID, videoIDs []uuid.UUID) (err error)
	Delete(ctx context.Context, userID uuid.UUID, videoID uuid.UUID) (err error)
	DeleteMany(ctx context.Context, userID uuid.UUID, videoIDs []uuid.UUID) (err error)
	DeleteAll(ctx context.Context, userID uuid.UUID) (err error)
	Get(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int64) (videos []uuid.UUID, err error)
	FanOut(ctx context.Context, channelID, videoID uuid.UUID) (err error)
}

func feedKey(userID uuid.UUID) string {
	return "feed:" + base64.RawURLEncoding.EncodeToString(userID[:])
}

// videoScore は video の UUIDv7 から score (ミリ秒) を取り出す。
// fan-out時のPushと同じscoreを再現できるので、後から createdAt を渡さずに push できる。
func videoScore(videoID uuid.UUID) float64 {
	return float64(util.TimeFromUUIDv7(videoID).UnixMilli())
}

type feedRepositoryImpl struct {
	client  redis.Cmdable
	maxSize int64
	q       sqlc.Querier
}

func NewFeedRepository(client redis.Cmdable, maxSize int64, q sqlc.Querier) FeedRepository {
	return &feedRepositoryImpl{
		client:  client,
		maxSize: maxSize,
		q:       q,
	}
}

func (f *feedRepositoryImpl) Push(ctx context.Context, userIDs []uuid.UUID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).Push(videoID=%s)", videoID)

	if len(userIDs) == 0 {
		return nil
	}

	score := videoScore(videoID)
	member := base64.RawURLEncoding.EncodeToString(videoID[:])

	pipe := f.client.Pipeline()
	for _, uid := range userIDs {
		key := feedKey(uid)
		pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
		// 古いエントリを切り詰めてメモリを一定に保つ
		pipe.ZRemRangeByRank(ctx, key, 0, -(f.maxSize + 1))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (f *feedRepositoryImpl) PushOne(ctx context.Context, userID uuid.UUID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).PushOne(userID=%s,videoID=%s)", userID, videoID)

	key := feedKey(userID)

	pipe := f.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: videoScore(videoID), Member: base64.RawURLEncoding.EncodeToString(videoID[:])})
	pipe.ZRemRangeByRank(ctx, key, 0, -(f.maxSize + 1))
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (f *feedRepositoryImpl) PushMany(ctx context.Context, userID uuid.UUID, videoIDs []uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).PushMany(userID=%s,count=%d)", userID, len(videoIDs))

	if len(videoIDs) == 0 {
		return nil
	}

	key := feedKey(userID)
	zs := make([]redis.Z, 0, len(videoIDs))
	for _, vid := range videoIDs {
		zs = append(zs, redis.Z{Score: videoScore(vid), Member: base64.RawURLEncoding.EncodeToString(vid[:])})
	}

	pipe := f.client.Pipeline()
	pipe.ZAdd(ctx, key, zs...)
	pipe.ZRemRangeByRank(ctx, key, 0, -(f.maxSize + 1))
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (f *feedRepositoryImpl) Delete(ctx context.Context, userID uuid.UUID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).Delete(userID=%s,videoID=%s)", userID, videoID)

	if err := f.client.ZRem(ctx, feedKey(userID), base64.RawURLEncoding.EncodeToString(videoID[:])).Err(); err != nil {
		return err
	}
	return nil
}

func (f *feedRepositoryImpl) DeleteAll(ctx context.Context, userID uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).DeleteAll(userID=%s)", userID)

	if err := f.client.Del(ctx, feedKey(userID)).Err(); err != nil {
		return err
	}
	return nil
}

func (f *feedRepositoryImpl) DeleteMany(ctx context.Context, userID uuid.UUID, videoIDs []uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).DeleteMany(userID=%s,count=%d)", userID, len(videoIDs))

	if len(videoIDs) == 0 {
		return nil
	}

	members := make([]interface{}, len(videoIDs))
	for i, vid := range videoIDs {
		members[i] = base64.RawURLEncoding.EncodeToString(vid[:])
	}
	if err := f.client.ZRem(ctx, feedKey(userID), members...).Err(); err != nil {
		return err
	}
	return nil
}

func (f *feedRepositoryImpl) Get(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int64) (_ []uuid.UUID, err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).Get(userID=%s)", userID)

	key := feedKey(userID)

	max := "+inf"
	if cursor != nil {
		cursorMember := base64.RawURLEncoding.EncodeToString(cursor[:])
		score, err := f.client.ZScore(ctx, key, cursorMember).Result()
		if errors.Is(err, redis.Nil) {
			// cursorがfeedに無い場合は空を返す
			return []uuid.UUID{}, nil
		}
		if err != nil {
			return nil, err
		}
		// 前ページ末尾のscoreより小さいものを取る（排他）
		max = "(" + strconv.FormatFloat(score, 'f', -1, 64)
	}

	members, err := f.client.ZRevRangeByScore(ctx, key, new(redis.ZRangeBy{
		Min:    "-inf",
		Max:    max,
		Offset: 0,
		Count:  limit,
	})).Result()
	if err != nil {
		return nil, err
	}

	videos := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		b, err := base64.RawURLEncoding.DecodeString(m)
		if err != nil {
			continue
		}
		videoID, err := uuid.FromBytes(b)
		if err != nil {
			continue
		}
		videos = append(videos, videoID)
	}
	return videos, nil
}

func (f *feedRepositoryImpl) FanOut(ctx context.Context, channelID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "database_d.(*feedRepositoryImpl).FanOut(videoID=%s)", videoID)

	subscribers, err := f.q.ListSubscribersByChannelPublicID(ctx, sqlc.ListSubscribersByChannelPublicIDParams{
		ChannelPublicID: channelID,
		VideoPublicID:   videoID,
	})
	if err != nil {
		return err
	}

	if err := f.Push(ctx, subscribers, videoID); err != nil {
		return err
	}
	return nil
}

var _ FeedRepository = (*feedRepositoryImpl)(nil)
