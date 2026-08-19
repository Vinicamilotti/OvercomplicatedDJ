package entities

import mediaEntities "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/media/entities"

type QueueEntry struct {
	Track     mediaEntities.Track
	Requester string
}

type Mode string

const (
	DJ  Mode = "DJ"
	RPG Mode = "RPG"
)