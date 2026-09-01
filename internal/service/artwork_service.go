package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ne/internal/domain"
	"ne/internal/media"
	"ne/internal/repository"
)

type ArtworkService struct {
	catalogRepo *repository.CatalogRepository
	cacheDir    string
}

func NewArtworkService(catalogRepo *repository.CatalogRepository, cacheDir string) *ArtworkService {
	artworkCache := filepath.Join(cacheDir, "artwork")
	_ = os.MkdirAll(artworkCache, 0755)
	return &ArtworkService{
		catalogRepo: catalogRepo,
		cacheDir:    artworkCache,
	}
}

// ServeArtwork streams artwork for a given album, track, or artist with disk and ETag caching.
func (s *ArtworkService) ServeArtwork(ctx context.Context, w http.ResponseWriter, r *http.Request, itemType, itemID string, size int) error {
	// 0. Check cache file first
	cacheFile := filepath.Join(s.cacheDir, fmt.Sprintf("%s_%s.jpg", itemType, itemID))
	if data, err := os.ReadFile(cacheFile); err == nil && len(data) > 0 {
		etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
		w.Header().Set("ETag", etag)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		http.ServeContent(w, r, "cover.jpg", time.Time{}, bytes.NewReader(data))
		return nil
	}

	var filePath string
	var searchArtist, searchTitle string

	switch itemType {
	case "track":
		track, err := s.catalogRepo.GetTrackByID(ctx, itemID)
		if err != nil {
			return err
		}
		filePath = track.Path
		searchArtist = track.RawArtist
		searchTitle = track.Title
	case "album":
		tracks, err := s.catalogRepo.ListTracksByAlbum(ctx, itemID)
		if err != nil || len(tracks) == 0 {
			return domain.ErrNotFound("Album artwork", itemID)
		}
		filePath = tracks[0].Path
		searchArtist = tracks[0].RawArtist
		searchTitle = tracks[0].AlbumTitle
	case "artist":
		artist, err := s.catalogRepo.GetArtistByID(ctx, itemID)
		if err != nil {
			return domain.ErrNotFound("Artist artwork", itemID)
		}
		searchArtist = artist.Name
		searchTitle = ""
		// If artist has albums, grab the first album's audio track for artwork extraction
		albums, _, _ := s.catalogRepo.ListAlbums(ctx, 100, 0)
		for _, alb := range albums {
			if alb.AlbumArtistID == itemID {
				albTracks, _ := s.catalogRepo.ListTracksByAlbum(ctx, alb.ID)
				if len(albTracks) > 0 {
					filePath = albTracks[0].Path
					break
				}
			}
		}
	default:
		return domain.ErrInvalidInput("Invalid artwork item type")
	}

	// 1. Try extracting local artwork
	var art *media.ExtractedArtwork
	var err error
	if filePath != "" {
		art, err = media.ExtractArtwork(filePath)
	} else {
		err = fmt.Errorf("no local audio file for artwork")
	}

	if err != nil {
		// 2. Fallback: Query online cover art provider (iTunes API)
		if onlineArt, onErr := media.FetchOnlineArtwork(searchArtist, searchTitle); onErr == nil {
			art = onlineArt
		} else {
			return s.servePlaceholderSVG(w, r)
		}
	}

	// Save to persistent cache
	_ = os.WriteFile(cacheFile, art.Data, 0644)

	etag := fmt.Sprintf(`"%s"`, art.Fingerprint)
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", art.MimeType)
	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	w.Header().Set("Content-Length", strconv.Itoa(len(art.Data)))

	http.ServeContent(w, r, "cover", time.Time{}, bytes.NewReader(art.Data))
	return nil
}

func (s *ArtworkService) servePlaceholderSVG(w http.ResponseWriter, r *http.Request) error {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="300" viewBox="0 0 300 300">
		<rect width="300" height="300" fill="#18181b"/>
		<circle cx="150" cy="150" r="100" fill="#27272a"/>
		<circle cx="150" cy="150" r="40" fill="#18181b"/>
		<circle cx="150" cy="150" r="15" fill="#f43f5e"/>
	</svg>`
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, err := w.Write([]byte(svg))
	return err
}
