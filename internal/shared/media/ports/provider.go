package ports

import "github.com/Vinicamilotti/OvercomplicatedDJ/internal/shared/media/entities"

type MediaProviderInterface interface {
	Search(query string) (*entities.Track, error)
	GetStreamURL(url string) (string, error)
	Supports(url string) bool
}