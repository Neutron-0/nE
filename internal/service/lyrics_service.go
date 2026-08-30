package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ne/internal/media"
	"ne/internal/repository"
)

type LyricsService struct {
	catalogRepo *repository.CatalogRepository
	cacheDir    string
}

func NewLyricsService(catalogRepo *repository.CatalogRepository, baseCacheDir string) *LyricsService {
	dir := filepath.Join(baseCacheDir, "lyrics")
	_ = os.MkdirAll(dir, 0750)
	return &LyricsService{
		catalogRepo: catalogRepo,
		cacheDir:    dir,
	}
}

func (s *LyricsService) GetLyrics(ctx context.Context, trackID string) (*media.LyricsResult, error) {
	// 1. Check cache file
	cachePath := filepath.Join(s.cacheDir, fmt.Sprintf("%s.json", trackID))
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached media.LyricsResult
		if err := json.Unmarshal(data, &cached); err == nil {
			return &cached, nil
		}
	}

	// 2. Fetch track info
	track, err := s.catalogRepo.GetTrackByID(ctx, trackID)
	if err != nil {
		return nil, err
	}

	result := &media.LyricsResult{
		TrackID: trackID,
		Source:  "none",
	}

	// 3. Check local .lrc file
	if lrcContent, err := media.FindLocalLRC(track.Path); err == nil {
		lines := media.ParseLRC(lrcContent)
		if len(lines) > 0 {
			result.IsSynced = true
			result.Lines = lines
			result.Source = "local"
		}
	}

	// 4. Fallback to LRCLIB public API
	if !result.IsSynced && result.PlainLyrics == "" {
		if lrclibRes, err := media.FetchLRCLIB(track.RawArtist, track.Title, track.AlbumTitle, track.Duration); err == nil {
			result.IsSynced = lrclibRes.IsSynced
			result.Lines = lrclibRes.Lines
			result.PlainLyrics = lrclibRes.PlainLyrics
			result.Source = "lrclib"
		}
	}

	// 5. Save to cache
	if result.IsSynced || result.PlainLyrics != "" {
		if data, err := json.Marshal(result); err == nil {
			_ = os.WriteFile(cachePath, data, 0644)
		}
	}

	return result, nil
}