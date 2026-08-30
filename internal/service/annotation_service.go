package service

import (
	"context"

	"ne/internal/domain"
	"ne/internal/repository"
)

type AnnotationService struct {
	repo *repository.AnnotationRepository
}

func NewAnnotationService(repo *repository.AnnotationRepository) *AnnotationService {
	return &AnnotationService{repo: repo}
}

func (s *AnnotationService) StarItem(ctx context.Context, userID, itemType, itemID string, starred bool) error {
	switch itemType {
	case "track":
		return s.repo.SetTrackStar(ctx, userID, itemID, starred)
	case "album":
		return s.repo.SetAlbumStar(ctx, userID, itemID, starred)
	case "artist":
		return s.repo.SetArtistStar(ctx, userID, itemID, starred)
	default:
		return domain.ErrInvalidInput("Invalid itemType (must be 'track', 'album', or 'artist')")
	}
}

func (s *AnnotationService) RateItem(ctx context.Context, userID, itemType, itemID string, rating int) error {
	switch itemType {
	case "track":
		return s.repo.SetTrackRating(ctx, userID, itemID, rating)
	case "album":
		return s.repo.SetAlbumRating(ctx, userID, itemID, rating)
	default:
		return domain.ErrInvalidInput("Invalid itemType (must be 'track' or 'album')")
	}
}

func (s *AnnotationService) Scrobble(ctx context.Context, userID, trackID, playerName string, durationPlayed float64, completed bool) error {
	return s.repo.RecordPlayback(ctx, userID, trackID, playerName, durationPlayed, completed)
}

func (s *AnnotationService) GetRecentHistory(ctx context.Context, userID string, limit int) ([]*domain.PlaybackRecord, error) {
	return s.repo.GetRecentHistory(ctx, userID, limit)
}

func (s *AnnotationService) GetFavorites(ctx context.Context, userID string) ([]*domain.Track, error) {
	return s.repo.GetFavoriteTracks(ctx, userID)
}
