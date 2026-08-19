package ports

import (
	"github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/queue/entities"
)

type QueueInterface interface {
	Add(entry entities.QueueEntry)
	Next() (entities.QueueEntry, bool)
	List() []entities.QueueEntry
	Clear()
	Len() int
}

type QueueServiceInterface interface {
	GetOrCreate(guildID string) QueueInterface
}