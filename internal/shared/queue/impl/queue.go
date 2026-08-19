package impl

import (
	"sync"

	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/entities"
)

type Queue struct {
	mu    sync.Mutex
	songs []entities.QueueEntry
}

func NewQueue() *Queue {
	return &Queue{songs: make([]entities.QueueEntry, 0)}
}

func (q *Queue) Add(entry entities.QueueEntry) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = append(q.songs, entry)
}

func (q *Queue) Next() (entities.QueueEntry, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.songs) == 0 {
		return entities.QueueEntry{}, false
	}
	entry := q.songs[0]
	q.songs = q.songs[1:]
	return entry, true
}

func (q *Queue) List() []entities.QueueEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	result := make([]entities.QueueEntry, len(q.songs))
	copy(result, q.songs)
	return result
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = make([]entities.QueueEntry, 0)
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.songs)
}