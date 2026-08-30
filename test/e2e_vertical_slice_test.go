package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestE2E_CompleteFullMVPArchitecture(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	configDir := filepath.Join(dir, "config")
	cacheDir := filepath.Join(dir, "cache")
	musicDir := filepath.Join(dir, "music")

	_ = os.MkdirAll(dataDir, 0750)
	_ = os.MkdirAll(configDir, 0750)
	_ = os.MkdirAll(cacheDir, 0750)

	// Create synthetic music library structure
	albumDir := filepath.Join(musicDir, "Pink Floyd", "Dark Side of the Moon")
	_ = os.MkdirAll(albumDir, 0755)

	track1Path := filepath.Join(albumDir, "01 - Speak to Me.flac")
	track2Path := filepath.Join(albumDir, "02 - Breathe.mp3")

	audioContent1 := []byte("FLAC_AUDIO_PAYLOAD_TRACK_1_SPEAK_TO_ME_SAMPLE_DATA_1234567890")
	audioContent2 := []byte("MP3_AUDIO_PAYLOAD_TRACK_2_BREATHE_SAMPLE_DATA_1234567890")

	_ = os.WriteFile(track1Path, audioContent1, 0644)
	_ = os.WriteFile(track2Path, audioContent2, 0644)

	// Setup full application environment
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = dataDir
	cfg.Paths.ConfigDir = configDir
	cfg.Paths.CacheDir = cacheDir
	cfg.Paths.MusicDir = musicDir

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := repository.Open(filepath.Join(dataDir, "ne.db"), 5000, logger)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	jwtMgr, err := auth.NewJWTManager("test-super-secret-encryption-key-32b!", configDir, 15*time.Minute, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("creating jwt manager: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	catalogRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)

	authService := service.NewAuthService(userRepo, jwtMgr)
	scannerEngine := scanner.NewScanner(db, catalogRepo, &cfg.Scanner, logger)
	catalogService := service.NewCatalogService(catalogRepo, scannerEngine)
	transcoderPool := media.NewTranscoder(2)
	streamService := service.NewStreamService(catalogRepo, transcoderPool)
	artworkService := service.NewArtworkService(catalogRepo, cacheDir)
	annoService := service.NewAnnotationService(annoRepo)
	playlistService := service.NewPlaylistService(playlistRepo)

	server := transport.NewServer(
		cfg,
		logger,
		func() error { return db.Ping(context.Background()) },
		jwtMgr,
		authService,
		catalogService,
		streamService,
		artworkService,
		annoService,
		playlistService,
	)
	router := server.Router()

	// STEP 1: First-Time Setup Wizard (Admin Registration)
	setupPayload, _ := json.Marshal(map[string]string{
		"username": "audiophile",
		"email":    "admin@ne.audio",
		"password": "audiophilepassword123",
	})
	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup", bytes.NewReader(setupPayload))
	setupRec := httptest.NewRecorder()
	router.ServeHTTP(setupRec, setupReq)

	if setupRec.Code != http.StatusCreated {
		t.Fatalf("Step 1 Failed: Setup returned status %d. Body: %s", setupRec.Code, setupRec.Body.String())
	}

	var authRes service.AuthResult
	if err := json.NewDecoder(setupRec.Body).Decode(&authRes); err != nil {
		t.Fatalf("Step 1 Failed: decoding setup response: %v", err)
	}
	token := authRes.AccessToken
	if token == "" {
		t.Fatal("Step 1 Failed: Empty JWT access token")
	}

	// STEP 2: Configure Music Folder (Library Ingestion)
	libPayload, _ := json.Marshal(map[string]string{
		"name": "Vinyl Masters",
		"path": musicDir,
	})
	libReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/libraries", bytes.NewReader(libPayload))
	libReq.Header.Set("Authorization", "Bearer "+token)
	libRec := httptest.NewRecorder()
	router.ServeHTTP(libRec, libReq)

	if libRec.Code != http.StatusCreated {
		t.Fatalf("Step 2 Failed: Create library returned status %d. Body: %s", libRec.Code, libRec.Body.String())
	}

	var lib domain.Library
	if err := json.NewDecoder(libRec.Body).Decode(&lib); err != nil {
		t.Fatalf("Step 2 Failed: decoding library response: %v", err)
	}

	// STEP 3: Scan Library Synchronously
	scanProgress, err := scannerEngine.ScanLibrary(context.Background(), lib.ID)
	if err != nil {
		t.Fatalf("Step 3 Failed: Library scan failed: %v", err)
	}
	if scanProgress.FilesProcessed != 2 {
		t.Fatalf("Step 3 Failed: Expected 2 files processed, got %d", scanProgress.FilesProcessed)
	}

	// STEP 4: Browse Catalog (Albums & Tracks)
	albumReq := httptest.NewRequest(http.MethodGet, "/api/v1/albums", nil)
	albumReq.Header.Set("Authorization", "Bearer "+token)
	albumRec := httptest.NewRecorder()
	router.ServeHTTP(albumRec, albumReq)

	if albumRec.Code != http.StatusOK {
		t.Fatalf("Step 4 Failed: List albums returned status %d", albumRec.Code)
	}

	var albumCol transport.CollectionResponse[*domain.Album]
	_ = json.NewDecoder(albumRec.Body).Decode(&albumCol)
	if len(albumCol.Items) == 0 {
		t.Fatal("Step 4 Failed: No albums found in catalog after scan")
	}

	trackReq := httptest.NewRequest(http.MethodGet, "/api/v1/tracks", nil)
	trackReq.Header.Set("Authorization", "Bearer "+token)
	trackRec := httptest.NewRecorder()
	router.ServeHTTP(trackRec, trackReq)

	if trackRec.Code != http.StatusOK {
		t.Fatalf("Step 4 Failed: List tracks returned status %d", trackRec.Code)
	}

	var trackCol transport.CollectionResponse[*domain.Track]
	_ = json.NewDecoder(trackRec.Body).Decode(&trackCol)
	if len(trackCol.Items) != 2 {
		t.Fatalf("Step 4 Failed: Expected 2 tracks in catalog, got %d", len(trackCol.Items))
	}

	targetTrack := trackCol.Items[0]

	// STEP 5: Search via FTS5 Index
	searchReq := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=Speak", nil)
	searchReq.Header.Set("Authorization", "Bearer "+token)
	searchRec := httptest.NewRecorder()
	router.ServeHTTP(searchRec, searchReq)

	if searchRec.Code != http.StatusOK {
		t.Fatalf("Step 5 Failed: Search returned status %d", searchRec.Code)
	}
	var searchRes domain.SearchResult
	_ = json.NewDecoder(searchRec.Body).Decode(&searchRes)
	if len(searchRes.Tracks) == 0 {
		t.Fatal("Step 5 Failed: FTS5 search did not find track 'Speak to Me'")
	}

	// STEP 6: Star Track & Retrieve Favorites
	starPayload, _ := json.Marshal(map[string]any{
		"itemType":  "track",
		"itemId":    targetTrack.ID,
		"isStarred": true,
	})
	starReq := httptest.NewRequest(http.MethodPost, "/api/v1/annotations/star", bytes.NewReader(starPayload))
	starReq.Header.Set("Authorization", "Bearer "+token)
	starRec := httptest.NewRecorder()
	router.ServeHTTP(starRec, starReq)

	if starRec.Code != http.StatusOK {
		t.Fatalf("Step 6 Failed: Star track returned status %d", starRec.Code)
	}

	favReq := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
	favReq.Header.Set("Authorization", "Bearer "+token)
	favRec := httptest.NewRecorder()
	router.ServeHTTP(favRec, favReq)

	var favCol transport.CollectionResponse[*domain.Track]
	_ = json.NewDecoder(favRec.Body).Decode(&favCol)
	if len(favCol.Items) != 1 || favCol.Items[0].ID != targetTrack.ID {
		t.Fatalf("Step 6 Failed: Expected favorite track %s", targetTrack.ID)
	}

	// STEP 7: Scrobble Playback
	scrobblePayload, _ := json.Marshal(map[string]any{
		"trackId":        targetTrack.ID,
		"playerName":     "Integration Test Runner",
		"durationPlayed": 237.0,
		"completed":      true,
	})
	scrobbleReq := httptest.NewRequest(http.MethodPost, "/api/v1/playback/scrobble", bytes.NewReader(scrobblePayload))
	scrobbleReq.Header.Set("Authorization", "Bearer "+token)
	scrobbleRec := httptest.NewRecorder()
	router.ServeHTTP(scrobbleRec, scrobbleReq)

	if scrobbleRec.Code != http.StatusCreated {
		t.Fatalf("Step 7 Failed: Scrobble returned status %d", scrobbleRec.Code)
	}

	// STEP 8: Create Playlist & Add Tracks
	plPayload, _ := json.Marshal(map[string]any{
		"name":     "Progressive Rock Essentials",
		"comment":  "Master tracks",
		"isPublic": true,
	})
	plReq := httptest.NewRequest(http.MethodPost, "/api/v1/playlists", bytes.NewReader(plPayload))
	plReq.Header.Set("Authorization", "Bearer "+token)
	plRec := httptest.NewRecorder()
	router.ServeHTTP(plRec, plReq)

	if plRec.Code != http.StatusCreated {
		t.Fatalf("Step 8 Failed: Create playlist returned status %d", plRec.Code)
	}
	var createdPL domain.Playlist
	_ = json.NewDecoder(plRec.Body).Decode(&createdPL)

	plTracksPayload, _ := json.Marshal(map[string][]string{
		"trackIds": {targetTrack.ID},
	})
	plTracksReq := httptest.NewRequest(http.MethodPut, "/api/v1/playlists/"+createdPL.ID+"/tracks", bytes.NewReader(plTracksPayload))
	plTracksReq.Header.Set("Authorization", "Bearer "+token)
	plTracksRec := httptest.NewRecorder()
	router.ServeHTTP(plTracksRec, plTracksReq)

	if plTracksRec.Code != http.StatusOK {
		t.Fatalf("Step 8 Failed: Add tracks to playlist returned status %d", plTracksRec.Code)
	}

	// STEP 9: Audio Stream & Seeking (HTTP 206 Partial Content Range Request)
	streamReq := httptest.NewRequest(http.MethodGet, "/api/v1/stream/"+targetTrack.ID, nil)
	streamReq.Header.Set("Authorization", "Bearer "+token)
	streamReq.Header.Set("Range", "bytes=0-19")
	streamRec := httptest.NewRecorder()
	router.ServeHTTP(streamRec, streamReq)

	if streamRec.Code != http.StatusPartialContent {
		t.Fatalf("Step 9 Failed: Expected 206 Partial Content for Range stream, got %d. Body: %s", streamRec.Code, streamRec.Body.String())
	}
	if streamRec.Body.Len() != 20 {
		t.Fatalf("Step 9 Failed: Expected 20 bytes streamed, got %d", streamRec.Body.Len())
	}

	// STEP 10: Artwork Endpoint
	artReq := httptest.NewRequest(http.MethodGet, "/api/v1/artwork/album/"+albumCol.Items[0].ID, nil)
	artRec := httptest.NewRecorder()
	router.ServeHTTP(artRec, artReq)

	if artRec.Code != http.StatusOK {
		t.Fatalf("Step 10 Failed: Artwork returned status %d", artRec.Code)
	}

	// STEP 11: Admin Stats
	statsReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	statsReq.Header.Set("Authorization", "Bearer "+token)
	statsRec := httptest.NewRecorder()
	router.ServeHTTP(statsRec, statsReq)

	if statsRec.Code != http.StatusOK {
		t.Fatalf("Step 11 Failed: Admin stats returned status %d", statsRec.Code)
	}
	var stats repository.SystemStats
	_ = json.NewDecoder(statsRec.Body).Decode(&stats)
	if stats.TotalTracks != 2 || stats.TotalPlaylists != 1 {
		t.Fatalf("Step 11 Failed: Stats mismatch. Tracks: %d, Playlists: %d", stats.TotalTracks, stats.TotalPlaylists)
	}

	// STEP 12: Serve Embedded Single-Page Application
	spaReq := httptest.NewRequest(http.MethodGet, "/", nil)
	spaRec := httptest.NewRecorder()
	router.ServeHTTP(spaRec, spaReq)

	if spaRec.Code != http.StatusOK {
		t.Fatalf("Step 12 Failed: SPA root returned status %d", spaRec.Code)
	}
}
