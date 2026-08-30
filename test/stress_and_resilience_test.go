package test_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ne/internal/auth"
	"ne/internal/config"
	"ne/internal/domain"
	"ne/internal/media"
	"ne/internal/repository"
	"ne/internal/scanner"
	"ne/internal/service"
	transport "ne/internal/transport/http"
)

func setupStressEnvironment(t *testing.T) (
	*transport.Server,
	*service.CatalogService,
	*service.AnnotationService,
	*service.PlaylistService,
	*repository.CatalogRepository,
	*repository.DB,
	string,
	string,
) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = dir
	cfg.Paths.CacheDir = filepath.Join(dir, "cache")
	cfg.Paths.ConfigDir = filepath.Join(dir, "config")
	_ = os.MkdirAll(cfg.Paths.DataDir, 0750)
	_ = os.MkdirAll(cfg.Paths.ConfigDir, 0750)
	_ = os.MkdirAll(cfg.Paths.CacheDir, 0750)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := repository.Open(filepath.Join(dir, "stress.db"), 5000, logger)
	if err != nil {
		t.Fatalf("failed to open stress db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	plRepo := repository.NewPlaylistRepository(db)

	jwtMgr, _ := auth.NewJWTManager("super-secure-test-jwt-secret-key-32-chars!", cfg.Paths.ConfigDir, 15*time.Minute, 30*24*time.Hour)
	authService := service.NewAuthService(userRepo, jwtMgr)
	scannerEngine := scanner.NewScanner(db, catRepo, &cfg.Scanner, logger)
	catService := service.NewCatalogService(catRepo, scannerEngine)
	transcoderPool := media.NewTranscoder(4)
	streamService := service.NewStreamService(catRepo, transcoderPool)
	artService := service.NewArtworkService(catRepo, cfg.Paths.CacheDir)
	annoService := service.NewAnnotationService(annoRepo)
	plService := service.NewPlaylistService(plRepo)

	srv := transport.NewServer(
		cfg,
		logger,
		func() error { return nil },
		jwtMgr,
		authService,
		catService,
		streamService,
		artService,
		annoService,
		plService,
		nil,
		nil,
		nil,
		nil,
	)

	// Create Admin User
	ctx := context.Background()
	adminAuth, err := authService.SetupInitialAdmin(ctx, "admin", "admin@ne.audio", "adminpassword123")
	if err != nil {
		t.Fatalf("setup admin failed: %v", err)
	}

	return srv, catService, annoService, plService, catRepo, db, adminAuth.AccessToken, dir
}

func TestStress_250AudioFiles_IngestionAndFTS(t *testing.T) {
	srv, _, _, _, catRepo, db, token, dir := setupStressEnvironment(t)
	router := srv.Router()
	ctx := context.Background()

	musicDir := filepath.Join(dir, "music")
	_ = os.MkdirAll(musicDir, 0755)

	// Create Library via API
	libBody, _ := json.Marshal(map[string]string{
		"name": "Stress Library",
		"path": musicDir,
	})
	libReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/libraries", bytes.NewReader(libBody))
	libReq.Header.Set("Authorization", "Bearer "+token)
	libRec := httptest.NewRecorder()
	router.ServeHTTP(libRec, libReq)

	if libRec.Code != http.StatusCreated {
		t.Fatalf("create library failed: %d", libRec.Code)
	}
	var createdLib domain.Library
	_ = json.NewDecoder(libRec.Body).Decode(&createdLib)

	// Generate 250 files across 10 artists and 25 albums
	artistNames := []string{
		"Björk", "Sigur Rós", "Daft Punk", "Massive Attack", "Boards of Canada",
		"Aphex Twin", "Kraftwerk", "Radiohead", "Portishead", "Four Tet",
	}

	totalTracks := 250
	var trackIDs []string

	for i := 0; i < totalTracks; i++ {
		artistName := artistNames[i%len(artistNames)]
		albumName := fmt.Sprintf("%s Album %d", artistName, (i/10)+1)
		trackTitle := fmt.Sprintf("Track %03d - Melodic Movement", i+1)

		art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), artistName, artistName, "")
		alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{
			Title:         albumName,
			SortTitle:     albumName,
			AlbumArtistID: art.ID,
			Year:          2000 + (i % 24),
		})

		trackFilePath := filepath.Join(musicDir, fmt.Sprintf("art_%02d_alb_%02d_trk_%03d.flac", i%10, i/10, i))
		_ = os.WriteFile(trackFilePath, []byte(fmt.Sprintf("FLAC DUMMY DATA FOR TRACK %d", i)), 0644)

		track := &domain.Track{
			ID:          fmt.Sprintf("stress-trk-%03d", i),
			PID:         fmt.Sprintf("stress-pid-%03d", i),
			LibraryID:   createdLib.ID,
			Path:        trackFilePath,
			FolderPath:  musicDir,
			Filename:    filepath.Base(trackFilePath),
			Title:       trackTitle,
			SortTitle:   trackTitle,
			RawArtist:   artistName,
			AlbumID:     alb.ID,
			Duration:    180.0 + float64(i),
			Format:      "flac",
			Codec:       "flac",
			FileSize:    1024,
			TrackNumber: (i % 10) + 1,
			DiscNumber:  1,
			MTime:       time.Now().Unix(),
		}
		_ = catRepo.UpsertTrack(ctx, db.Executor(), track)
		_ = catRepo.IndexTrackFTS(ctx, db.Executor(), track.ID, track.Title, track.RawArtist, alb.Title, "Electronic")
		trackIDs = append(trackIDs, track.ID)
	}

	// 1. Verify Catalog Counts
	albums, _, err := catRepo.ListAlbums(ctx, 100, 0)
	if err != nil {
		t.Fatalf("list albums failed: %v", err)
	}
	if len(albums) < 25 {
		t.Errorf("expected at least 25 albums, got %d", len(albums))
	}

	// 2. Full-Text Search Stress Under Concurrent Query Load
	var wg sync.WaitGroup
	var searchSuccess atomic.Int64

	for w := 0; w < 20; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			queryArtist := artistNames[workerID%len(artistNames)]
			searchReq := httptest.NewRequest(http.MethodGet, "/api/v1/search?q="+url.QueryEscape(queryArtist), nil)
			searchReq.Header.Set("Authorization", "Bearer "+token)
			searchRec := httptest.NewRecorder()
			router.ServeHTTP(searchRec, searchReq)

			if searchRec.Code == http.StatusOK {
				var res domain.SearchResult
				if err := json.NewDecoder(searchRec.Body).Decode(&res); err == nil && len(res.Tracks) > 0 {
					searchSuccess.Add(1)
				}
			}
		}(w)
	}
	wg.Wait()

	if searchSuccess.Load() != 20 {
		t.Errorf("expected 20 successful concurrent search requests, got %d", searchSuccess.Load())
	}

	// 3. Concurrent Range Streaming & Playback Scrobbling
	var streamSuccess atomic.Int64
	for s := 0; s < 30; s++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			targetTrackID := trackIDs[idx%len(trackIDs)]

			// Range Stream Request
			sReq := httptest.NewRequest(http.MethodGet, "/api/v1/stream/"+targetTrackID, nil)
			sReq.Header.Set("Authorization", "Bearer "+token)
			sReq.Header.Set("Range", "bytes=0-20")
			sRec := httptest.NewRecorder()
			router.ServeHTTP(sRec, sReq)

			if sRec.Code == http.StatusPartialContent && sRec.Body.Len() == 21 {
				streamSuccess.Add(1)
			}

			// Concurrent Scrobble
			scrobBody, _ := json.Marshal(map[string]any{
				"trackId":        targetTrackID,
				"durationPlayed": 180.0,
				"completed":      true,
			})
			scrobReq := httptest.NewRequest(http.MethodPost, "/api/v1/playback/scrobble", bytes.NewReader(scrobBody))
			scrobReq.Header.Set("Authorization", "Bearer "+token)
			scrobRec := httptest.NewRecorder()
			router.ServeHTTP(scrobRec, scrobReq)
		}(s)
	}
	wg.Wait()

	if streamSuccess.Load() != 30 {
		t.Errorf("expected 30 successful concurrent stream requests, got %d", streamSuccess.Load())
	}
}
