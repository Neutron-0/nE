package service

import (
	"context"

	"ne/internal/domain"
	"ne/internal/repository"
)

type SmartPlaylistService struct {
	catalogRepo *repository.CatalogRepository
}

func NewSmartPlaylistService(catalogRepo *repository.CatalogRepository) *SmartPlaylistService {
	return &SmartPlaylistService{catalogRepo: catalogRepo}
}

type SmartPlaylistSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func (s *SmartPlaylistService) ListAvailableSmartPlaylists() []SmartPlaylistSummary {
	return []SmartPlaylistSummary{
		{ID: "recently-added", Name: "Recently Added", Description: "Newest tracks scanned into the library", Icon: "Clock"},
		{ID: "most-played", Name: "Most Played", Description: "Your most frequently played tracks", Icon: "Flame"},
		{ID: "recently-played", Name: "Recently Played", Description: "Tracks you've listened to recently", Icon: "History"},
		{ID: "top-rated", Name: "Top Rated", Description: "Tracks rated 4 or 5 stars", Icon: "Star"},
		{ID: "random-mix", Name: "Random Mix", Description: "Shuffle of tracks across your library", Icon: "Shuffle"},
	}
}

func (s *SmartPlaylistService) GetSmartPlaylistTracks(ctx context.Context, userID, playlistID string, limit int) ([]*domain.Track, error) {
	if limit <= 0 {
		limit = 50
	}
	switch playlistID {
	case "recently-added":
		return s.catalogRepo.GetRecentlyAddedTracks(ctx, limit)
	case "most-played":
		return s.catalogRepo.GetMostPlayedTracks(ctx, userID, limit)
	case "recently-played":
		return s.catalogRepo.GetRecentlyPlayedTracks(ctx, userID, limit)
	case "top-rated":
		return s.catalogRepo.GetTopRatedTracks(ctx, userID, limit)
	case "random-mix":
		return s.catalogRepo.GetRandomTracks(ctx, limit)
	default:
		return s.catalogRepo.GetRecentlyAddedTracks(ctx, limit)
	}
}
