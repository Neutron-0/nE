package http_test

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

func newTestServer(t *testing.T, dbChecker func() error) (
	*transport.Server,
	*service.AuthService,
	*service.CatalogService,
	*service.StreamService,
	*service.ArtworkService,
	*service.AnnotationService,
	*service.PlaylistService,
	*repository.CatalogRepository,
	*repository.DB,
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
	db, err := repository.Open(filepath.Join(dir, "test.db"), 5000, logger)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	plRepo := repository.NewPlaylistRepository(db)

	jwtMgr, err := auth.NewJWTManager("test-secret-at-least-32-chars-long!", cfg.Paths.ConfigDir, 15*time.Minute, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	authService := service.NewAuthService(userRepo, jwtMgr)
	scannerEngine := scanner.NewScanner(db, catRepo, &cfg.Scanner, logger)
	catService := service.NewCatalogService(catRepo, scannerEngine)
	transcoderPool := media.NewTranscoder(2)
	streamService := service.NewStreamService(catRepo, transcoderPool)
	artService := service.NewArtworkService(catRepo, cfg.Paths.CacheDir)
	annoService := service.NewAnnotationService(annoRepo)
	plService := service.NewPlaylistService(plRepo)

	srv := transport.NewServer(
		cfg,
		logger,
		dbChecker,
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

	return srv, authService, catService, streamService, artService, annoService, plService, catRepo, db, dir
}

func TestCompleteHTTPServerEndpoints(t *testing.T) {
	srv, _, _, _, _, _, _, catRepo, db, dir := newTestServer(t, func() error { return nil })
	router := srv.Router()
	ctx := context.Background()

	// 1. Health Endpoints
	t.Run("Health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health/liveness", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("liveness code %d != 200", rec.Code)
		}

		diagReq := httptest.NewRequest(http.MethodGet, "/api/v1/health/diagnostics", nil)
		diagRec := httptest.NewRecorder()
		router.ServeHTTP(diagRec, diagReq)
		if diagRec.Code != http.StatusOK {
			t.Errorf("diagnostics code %d != 200", diagRec.Code)
		}
	})

	// 2. Auth: First-time Admin Setup
	setupBody := []byte(`{"username":"admin","email":"admin@ne.audio","password":"securepassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("setup failed with code %d", rec.Code)
	}

	var setupRes service.AuthResult
	_ = json.NewDecoder(rec.Body).Decode(&setupRes)
	token := setupRes.AccessToken

	// 3. Admin: Add Library
	musicDir := filepath.Join(dir, "music")
	_ = os.MkdirAll(musicDir, 0755)
	audioFilePath := filepath.Join(musicDir, "song.flac")
	audioData := []byte("FLAC HEADER + 1024 BYTES OF RAW TEST AUDIO DATA FOR STREAMING VERIFICATION")
	_ = os.WriteFile(audioFilePath, audioData, 0644)

	// Create cover art in folder
	coverPath := filepath.Join(musicDir, "cover.jpg")
	_ = os.WriteFile(coverPath, []byte("JPEG_COVER_ART_BYTES"), 0644)

	libBody, _ := json.Marshal(map[string]string{
		"name": "Local Music",
		"path": musicDir,
	})
	libReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/libraries", bytes.NewReader(libBody))
	libReq.Header.Set("Authorization", "Bearer "+token)
	libRec := httptest.NewRecorder()
	router.ServeHTTP(libRec, libReq)

	if libRec.Code != http.StatusCreated {
		t.Fatalf("create library failed with code %d. Body: %s", libRec.Code, libRec.Body.String())
	}

	var createdLib domain.Library
	_ = json.NewDecoder(libRec.Body).Decode(&createdLib)

	// 4. Ingest Track into DB with valid LibraryID
	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Radiohead", "Radiohead", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "In Rainbows", SortTitle: "In Rainbows", AlbumArtistID: art.ID})

	track := &domain.Track{
		ID:          "trk-1",
		PID:         "pid-1",
		LibraryID:   createdLib.ID,
		Path:        audioFilePath,
		FolderPath:  musicDir,
		Filename:    "song.flac",
		Title:       "15 Step",
		SortTitle:   "15 Step",
		RawArtist:   "Radiohead",
		AlbumID:     alb.ID,
		Duration:    237.0,
		Format:      "flac",
		Codec:       "flac",
		FileSize:    int64(len(audioData)),
		TrackNumber: 1,
		DiscNumber:  1,
		MTime:       time.Now().Unix(),
	}
	if err := catRepo.UpsertTrack(ctx, db.Executor(), track); err != nil {
		t.Fatalf("upsert track failed: %v", err)
	}
	_ = catRepo.IndexTrackFTS(ctx, db.Executor(), track.ID, track.Title, track.RawArtist, alb.Title, "Alternative")

	// 5. Catalog APIs: Albums, Artists, Tracks, Search
	t.Run("Catalog", func(t *testing.T) {
		// List Albums
		r1 := httptest.NewRequest(http.MethodGet, "/api/v1/albums", nil)
		r1.Header.Set("Authorization", "Bearer "+token)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, r1)
		if w1.Code != http.StatusOK {
			t.Errorf("list albums code %d != 200", w1.Code)
		}

		// Search FTS5
		r2 := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=Radiohead", nil)
		r2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, r2)
		if w2.Code != http.StatusOK {
			t.Errorf("search code %d != 200", w2.Code)
		}
	})

	// 6. Direct Stream & Range Seeking
	t.Run("Stream", func(t *testing.T) {
		rangeReq := httptest.NewRequest(http.MethodGet, "/api/v1/stream/trk-1", nil)
		rangeReq.Header.Set("Authorization", "Bearer "+token)
		rangeReq.Header.Set("Range", "bytes=0-10")
		rangeRec := httptest.NewRecorder()
		router.ServeHTTP(rangeRec, rangeReq)

		if rangeRec.Code != http.StatusPartialContent {
			t.Fatalf("expected 206 Partial Content, got %d (body: %s)", rangeRec.Code, rangeRec.Body.String())
		}
		if rangeRec.Body.Len() != 11 {
			t.Errorf("expected 11 bytes, got %d", rangeRec.Body.Len())
		}
	})

	// 7. Artwork Delivery
	t.Run("Artwork", func(t *testing.T) {
		artReq := httptest.NewRequest(http.MethodGet, "/api/v1/artwork/album/"+alb.ID+"?token="+token, nil)
		artRec := httptest.NewRecorder()
		router.ServeHTTP(artRec, artReq)

		if artRec.Code != http.StatusOK {
			t.Errorf("artwork delivery failed with status %d (body: %s)", artRec.Code, artRec.Body.String())
		}
	})

	// 8. Annotations (Star, Scrobble, Favorites, History)
	t.Run("AnnotationsAndScrobble", func(t *testing.T) {
		// Star Track
		starBody := []byte(`{"itemType":"track","itemId":"trk-1","isStarred":true}`)
		rStar := httptest.NewRequest(http.MethodPost, "/api/v1/annotations/star", bytes.NewReader(starBody))
		rStar.Header.Set("Authorization", "Bearer "+token)
		wStar := httptest.NewRecorder()
		router.ServeHTTP(wStar, rStar)
		if wStar.Code != http.StatusOK {
			t.Errorf("star track status %d != 200 (body: %s)", wStar.Code, wStar.Body.String())
		}

		// Get Favorites
		rFav := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
		rFav.Header.Set("Authorization", "Bearer "+token)
		wFav := httptest.NewRecorder()
		router.ServeHTTP(wFav, rFav)
		if wFav.Code != http.StatusOK {
			t.Errorf("get favorites status %d != 200 (body: %s)", wFav.Code, wFav.Body.String())
		}

		// Scrobble
		scrobBody := []byte(`{"trackId":"trk-1","durationPlayed":237.0,"completed":true}`)
		rScrob := httptest.NewRequest(http.MethodPost, "/api/v1/playback/scrobble", bytes.NewReader(scrobBody))
		rScrob.Header.Set("Authorization", "Bearer "+token)
		wScrob := httptest.NewRecorder()
		router.ServeHTTP(wScrob, rScrob)
		if wScrob.Code != http.StatusCreated {
			t.Errorf("scrobble status %d != 201 (body: %s)", wScrob.Code, wScrob.Body.String())
		}

		// Get History
		rHist := httptest.NewRequest(http.MethodGet, "/api/v1/history/recent", nil)
		rHist.Header.Set("Authorization", "Bearer "+token)
		wHist := httptest.NewRecorder()
		router.ServeHTTP(wHist, rHist)
		if wHist.Code != http.StatusOK {
			t.Errorf("get history status %d != 200 (body: %s)", wHist.Code, wHist.Body.String())
		}
	})

	// 9. Playlists
	t.Run("Playlists", func(t *testing.T) {
		// Create Playlist
		plBody := []byte(`{"name":"My Test Playlist","comment":"Cool beats","isPublic":false}`)
		rCreate := httptest.NewRequest(http.MethodPost, "/api/v1/playlists", bytes.NewReader(plBody))
		rCreate.Header.Set("Authorization", "Bearer "+token)
		wCreate := httptest.NewRecorder()
		router.ServeHTTP(wCreate, rCreate)
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("create playlist status %d != 201 (body: %s)", wCreate.Code, wCreate.Body.String())
		}

		var createdPl domain.Playlist
		_ = json.NewDecoder(wCreate.Body).Decode(&createdPl)

		// Set Playlist Tracks
		tracksBody := []byte(`{"trackIds":["trk-1"]}`)
		rSet := httptest.NewRequest(http.MethodPut, "/api/v1/playlists/"+createdPl.ID+"/tracks", bytes.NewReader(tracksBody))
		rSet.Header.Set("Authorization", "Bearer "+token)
		wSet := httptest.NewRecorder()
		router.ServeHTTP(wSet, rSet)
		if wSet.Code != http.StatusOK {
			t.Errorf("set playlist tracks status %d != 200 (body: %s)", wSet.Code, wSet.Body.String())
		}

		// List Playlists
		rList := httptest.NewRequest(http.MethodGet, "/api/v1/playlists", nil)
		rList.Header.Set("Authorization", "Bearer "+token)
		wList := httptest.NewRecorder()
		router.ServeHTTP(wList, rList)
		if wList.Code != http.StatusOK {
			t.Errorf("list playlists status %d != 200 (body: %s)", wList.Code, wList.Body.String())
		}
	})

	// 10. Security Checks (Unauthorized & Non-Admin)
	t.Run("SecurityEnforcement", func(t *testing.T) {
		// Unauthenticated catalog access -> 401
		unauthReq := httptest.NewRequest(http.MethodGet, "/api/v1/tracks", nil)
		unauthRec := httptest.NewRecorder()
		router.ServeHTTP(unauthRec, unauthReq)
		if unauthRec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for unauth request, got %d", unauthRec.Code)
		}
	})
}
