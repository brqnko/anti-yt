package database_d

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

type jtiEntry struct {
	jti       uuid.UUID
	expiresAt time.Time
}

type jtiEntryHeap []jtiEntry

func (h jtiEntryHeap) Len() int            { return len(h) }
func (h jtiEntryHeap) Less(i, j int) bool  { return h[i].expiresAt.Before(h[j].expiresAt) }
func (h jtiEntryHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *jtiEntryHeap) Push(x any)         { *h = append(*h, x.(jtiEntry)) }
func (h *jtiEntryHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

type jtiBlacklistRepositoryInMemory struct {
	mu      sync.Mutex
	entries map[uuid.UUID]time.Time
	queue   jtiEntryHeap
}

func (j *jtiBlacklistRepositoryInMemory) InsertJTI(ctx context.Context, jti uuid.UUID, expiresAt time.Time) (err error) {
	defer util.Wrap(&err, "database_d.(*jtiBlacklistRepositoryInMemory).InsertJTI(jti=%s)", jti)

	if time.Until(expiresAt) <= 0 {
		return nil
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	now := time.Now()
	for j.queue.Len() > 0 && !j.queue[0].expiresAt.After(now) {
		top := heap.Pop(&j.queue).(jtiEntry)
		if exp, ok := j.entries[top.jti]; ok && !exp.After(top.expiresAt) {
			delete(j.entries, top.jti)
		}
	}

	j.entries[jti] = expiresAt
	heap.Push(&j.queue, jtiEntry{jti: jti, expiresAt: expiresAt})
	return nil
}

func (j *jtiBlacklistRepositoryInMemory) IsJtiExist(ctx context.Context, jti uuid.UUID) (found bool, err error) {
	defer util.Wrap(&err, "database_d.(*jtiBlacklistRepositoryInMemory).IsJtiExist(jti=%s)", jti)

	j.mu.Lock()
	defer j.mu.Unlock()
	_, ok := j.entries[jti]
	return ok, nil
}

func NewInMemoryJtiBlacklistRepository() JtiBlacklistRepository {
	return &jtiBlacklistRepositoryInMemory{
		entries: make(map[uuid.UUID]time.Time),
	}
}

var _ JtiBlacklistRepository = (*jtiBlacklistRepositoryInMemory)(nil)
