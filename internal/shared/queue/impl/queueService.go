package impl

import (
	"sync"

	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/ports"
)

type QueueService struct {
	mu     sync.RWMutex
	queues map[string]ports.QueueInterface
}

func NewQueueService() *QueueService {
	return &QueueService{queues: make(map[string]ports.QueueInterface)}
}

func (qs *QueueService) GetOrCreate(guildID string) ports.QueueInterface {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	if q, ok := qs.queues[guildID]; ok {
		return q
	}
	q := NewQueue()
	qs.queues[guildID] = q
	return q
}