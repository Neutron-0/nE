package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ne/internal/domain"
	"ne/internal/repository"
	"ne/internal/scanner"
	"github.com/google/uuid"
)

type CatalogService struct {
	catalogRepo *repository.CatalogRepository
	scanner     *scanner.Scanner
}

func NewCatalogService(catalogRepo *repository.CatalogRepository, scanner *scanner.Scanner) *CatalogService {
	return &CatalogService{
		catalogRepo: catalogRepo,
		scanner:     scanner,
	}
}

// ----------------- Artists -----------------

func (s *CatalogService) ListArtists(ctx context.Context, limit, offset int) ([]*domain.Artist, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return s.catalogRepo.ListArtists(ctx, limit, offset)
}

func (s *CatalogService) GetArtist(ctx context.Context, id string) (*domain.Artist, error) {
	return s.catalogRepo.GetArtistByID(ctx, id)
}

// ----------------- Albums -----------------

func (s *CatalogService) ListAlbums(ctx context.Context, limit, offset int) ([]*domain.Album, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return s.catalogRepo.ListAlbums(ctx, limit, offset)
}

func (s *CatalogService) GetAlbum(ctx context.Context, id string) (*domain.Album, error) {
	album, err := s.catalogRepo.GetAlbumByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tracks, err := s.catalogRepo.ListTracksByAlbum(ctx, id)
	if err == nil {
		album.Tracks = tracks
	}

	return album, nil
}

// ----------------- Tracks -----------------

func (s *CatalogService) ListTracks(ctx context.Context, limit, offset int) ([]*domain.Track, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return s.catalogRepo.ListTracks(ctx, limit, offset)
}

func (s *CatalogService) GetTrack(ctx context.Context, id string) (*domain.Track, error) {
	return s.catalogRepo.GetTrackByID(ctx, id)
}

// ----------------- Genres -----------------

func (s *CatalogService) ListGenres(ctx context.Context) ([]*domain.Genre, error) {
	return s.catalogRepo.ListGenres(ctx)
}

// ----------------- Full Text Search -----------------

func (s *CatalogService) Search(ctx context.Context, query string, limit int) (*domain.SearchResult, error) {
	return s.catalogRepo.Search(ctx, query, limit)
}

// ----------------- System Stats -----------------

func (s *CatalogService) GetStats(ctx context.Context) (*repository.SystemStats, error) {
	return s.catalogRepo.GetStats(ctx)
}

// ----------------- Libraries & Scanning -----------------

func (s *CatalogService) ListLibraries(ctx context.Context) ([]*domain.Library, error) {
	return s.catalogRepo.ListLibraries(ctx)
}

func (s *CatalogService) CreateLibrary(ctx context.Context, name, path string) (*domain.Library, error) {
	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)

	if name == "" {
		return nil, domain.ErrInvalidInput("Library name cannot be empty")
	}
	if path == "" {
		return nil, domain.ErrInvalidInput("Library path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, domain.ErrInvalidInput(fmt.Sprintf("Directory does not exist or is inaccessible: %q", cleanPath))
	}

	lib := &domain.Library{
		ID:         uuid.NewString(),
		Name:       name,
		Path:       cleanPath,
		ScanStatus: "idle",
	}

	if err := s.catalogRepo.CreateLibrary(ctx, lib); err != nil {
		return nil, fmt.Errorf("persisting library: %w", err)
	}

	return lib, nil
}

func (s *CatalogService) TriggerScan(ctx context.Context, libraryID string) (*scanner.ScanProgress, error) {
	lib, err := s.catalogRepo.GetLibraryByID(ctx, libraryID)
	if err != nil {
		return nil, err
	}

	if lib.ScanStatus == "scanning" {
		return nil, domain.ErrConflict("Scan is already in progress for this library")
	}

	// Trigger async scan in background
	go func() {
		_, _ = s.scanner.ScanLibrary(context.Background(), libraryID)
	}()

	return &scanner.ScanProgress{
		LibraryID: libraryID,
		Status:    "scanning",
	}, nil
}

// ScanLibrarySync triggers synchronous library scan and waits for completion.
func (s *CatalogService) ScanLibrarySync(ctx context.Context, libraryID string) (*scanner.ScanProgress, error) {
	return s.scanner.ScanLibrary(ctx, libraryID)
}
