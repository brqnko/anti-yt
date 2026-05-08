package testutil

import (
	"context"
	"sort"
	"sync"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

// FakeFeedRepository は Redis を使わない in-memory 版の FeedRepository。
// ZSET と同じく videoID から videoScore (UUIDv7 のミリ秒) を導出して順序を決める。
type FakeFeedRepository struct {
	mu    sync.Mutex
	feeds map[uuid.UUID]map[uuid.UUID]float64
	q     sqlc.Querier
}

func NewFakeFeedRepository(q sqlc.Querier) *FakeFeedRepository {
	return new(FakeFeedRepository{feeds: map[uuid.UUID]map[uuid.UUID]float64{}, q: q})
}

func (f *FakeFeedRepository) getOrInit(userID uuid.UUID) map[uuid.UUID]float64 {
	m, ok := f.feeds[userID]
	if !ok {
		m = map[uuid.UUID]float64{}
		f.feeds[userID] = m
	}
	return m
}

func videoScore(v uuid.UUID) float64 {
	return float64(util.TimeFromUUIDv7(v).UnixMilli())
}

func (f *FakeFeedRepository) Push(_ context.Context, userIDs []uuid.UUID, videoID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := videoScore(videoID)
	for _, uid := range userIDs {
		f.getOrInit(uid)[videoID] = s
	}
	return nil
}

func (f *FakeFeedRepository) PushOne(_ context.Context, userID uuid.UUID, videoID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getOrInit(userID)[videoID] = videoScore(videoID)
	return nil
}

func (f *FakeFeedRepository) PushMany(_ context.Context, userID uuid.UUID, videoIDs []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m := f.getOrInit(userID)
	for _, v := range videoIDs {
		m[v] = videoScore(v)
	}
	return nil
}

func (f *FakeFeedRepository) Delete(_ context.Context, userID uuid.UUID, videoID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m, ok := f.feeds[userID]; ok {
		delete(m, videoID)
	}
	return nil
}

func (f *FakeFeedRepository) DeleteMany(_ context.Context, userID uuid.UUID, videoIDs []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.feeds[userID]
	if !ok {
		return nil
	}
	for _, v := range videoIDs {
		delete(m, v)
	}
	return nil
}

func (f *FakeFeedRepository) DeleteAll(_ context.Context, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.feeds, userID)
	return nil
}

func (f *FakeFeedRepository) Get(_ context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int64) ([]uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.feeds[userID]
	if !ok {
		return []uuid.UUID{}, nil
	}

	type entry struct {
		id    uuid.UUID
		score float64
	}
	entries := make([]entry, 0, len(m))
	for id, s := range m {
		entries = append(entries, entry{id, s})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].score != entries[j].score {
			return entries[i].score > entries[j].score
		}
		return entries[i].id.String() > entries[j].id.String()
	})

	if cursor != nil {
		cs, ok := m[*cursor]
		if !ok {
			return []uuid.UUID{}, nil
		}
		filtered := entries[:0:0]
		for _, e := range entries {
			if e.score < cs {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	if int64(len(entries)) > limit {
		entries = entries[:limit]
	}
	out := make([]uuid.UUID, len(entries))
	for i, e := range entries {
		out[i] = e.id
	}
	return out, nil
}

// Count はテスト用にユーザーの feed サイズを返す。
func (f *FakeFeedRepository) Count(userID uuid.UUID) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.feeds[userID])
}

// Has はテスト用に videoID が含まれているかを返す。
func (f *FakeFeedRepository) Has(userID, videoID uuid.UUID) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.feeds[userID]
	if !ok {
		return false
	}
	_, ok = m[videoID]
	return ok
}

func (f *FakeFeedRepository) FanOut(ctx context.Context, channelID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "testutil.(*FakeFeedRepository).FanOut(videoID=%s)", videoID)

	if f.q == nil {
		return nil
	}
	subscribers, err := f.q.ListSubscribersByChannelPublicID(ctx, sqlc.ListSubscribersByChannelPublicIDParams{
		ChannelPublicID: channelID,
		VideoPublicID:   videoID,
	})
	if err != nil {
		return err
	}
	return f.Push(ctx, subscribers, videoID)
}

var _ database_d.FeedRepository = (*FakeFeedRepository)(nil)
