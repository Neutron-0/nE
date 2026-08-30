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
	"os"
	"os/exec"
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

// generateAudioFixture uses FFmpeg to synthesize real, valid encoded audio with metadata and optional embedded artwork.
func generateAudioFixture(t *testing.T, path string, format string, title, artist, album string, trackNum, year int, coverPath string) {
	t.Helper()
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	args := []string{
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration=1",
	}

	if coverPath != "" {
		args = append(args, "-i", coverPath, "-map", "0:0", "-map", "1:0", "-c:v", "copy", "-disposition:v", "attached_pic")
	}

	// Add metadata
	args = append(args,
		"-metadata", fmt.Sprintf("title=%s", title),
		"-metadata", fmt.Sprintf("artist=%s", artist),
		"-metadata", fmt.Sprintf("album=%s", album),
		"-metadata", fmt.Sprintf("track=%d", trackNum),
		"-metadata", fmt.Sprintf("date=%d", year),
		"-y", path,
	)

	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg generate audio failed (%s): %v\nOutput: %s", path, err, string(out))
	}
}

// generateImageFixture creates a real valid JPEG image.
func generateImageFixture(t *testing.T, path string, width, height int) {
	t.Helper()
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	args := []string{
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=red:s=%dx%d:d=1", width, height),
		"-frames:v", "1",
		"-y", path,
	}

	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg generate image failed (%s): %v\nOutput: %s", path, err, string(out))
	}
}

func setupRealServer(t *testing.T) (
	*transport.Server,
	*service.CatalogService,
	*repository.CatalogRepository,
	*repository.AnnotationRepository,
	*repository.PlaylistRepository,
	*repository.DB,
	string,
	string,
	string,
) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = filepath.Join(dir, "data")
	cfg.Paths.CacheDir = filepath.Join(dir, "cache")
	cfg.Paths.ConfigDir = filepath.Join(dir, "config")
	_ = os.MkdirAll(cfg.Paths.DataDir, 0750)
	_ = os.MkdirAll(cfg.Paths.ConfigDir, 0750)
	_ = os.MkdirAll(cfg.Paths.CacheDir, 0750)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := repository.Open(filepath.Join(cfg.Paths.DataDir, "ne_real.db"), 5000, logger)
	if err != nil {
		t.Fatalf("failed to open real test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	plRepo := repository.NewPlaylistRepository(db)

	jwtMgr, err := auth.NewJWTManager("super-secret-jwt-key-32-chars-long-validation!", cfg.Paths.ConfigDir, 15*time.Minute, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to initialize jwt: %v", err)
	}

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
		func() error { return db.Ping(context.Background()) },
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

	// Create root admin
	ctx := context.Background()
	adminRes, err := authService.SetupInitialAdmin(ctx, "admin", "admin@ne.audio", "securepassword123")
	if err != nil {
		t.Fatalf("setup admin failed: %v", err)
	}

	return srv, catService, catRepo, annoRepo, plRepo, db, adminRes.AccessToken, adminRes.User.ID, dir
}

func TestRealWorld_Ingestion_Metadata_Playback_Lifecycle(t *testing.T) {
	srv, catService, catRepo, annoRepo, plRepo, db, token, userID, dir := setupRealServer(t)
	router := srv.Router()
	ctx := context.Background()

	// 1. Create real filesystem music library
	musicDir := filepath.Join(dir, "music")
	albumDir := filepath.Join(musicDir, "Boards of Canada", "Music Has the Right to Children")
	_ = os.MkdirAll(albumDir, 0755)

	coverPath := filepath.Join(albumDir, "cover.jpg")
	generateImageFixture(t, coverPath, 400, 400)

	// Real audio formats
	flacTrack := filepath.Join(albumDir, "01 - Wildlife Analysis.flac")
	generateAudioFixture(t, flacTrack, "flac", "Wildlife Analysis", "Boards of Canada", "Music Has the Right to Children", 1, 1998, coverPath)

	mp3Track := filepath.Join(albumDir, "02 - An Eagle in Your Mind.mp3")
	generateAudioFixture(t, mp3Track, "mp3", "An Eagle in Your Mind", "Boards of Canada", "Music Has the Right to Children", 2, 1998, coverPath)

	oggTrack := filepath.Join(albumDir, "03 - The Color of the Fire.ogg")
	generateAudioFixture(t, oggTrack, "ogg", "The Color of the Fire", "Boards of Canada", "Music Has the Right to Children", 3, 1998, "")

	wavTrack := filepath.Join(albumDir, "04 - Telephasic Workshop.wav")
	generateAudioFixture(t, wavTrack, "wav", "Telephasic Workshop", "Boards of Canada", "Music Has the Right to Children", 4, 1998, "")

	// 2. Add Library via API & Trigger Real Ingestion
	libBody, _ := json.Marshal(map[string]string{
		"name": "Electronic Library",
		"path": musicDir,
	})
	libReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/libraries", bytes.NewReader(libBody))
	libReq.Header.Set("Authorization", "Bearer "+token)
	libRec := httptest.NewRecorder()
	router.ServeHTTP(libRec, libReq)

	if libRec.Code != http.StatusCreated {
		t.Fatalf("create library failed: %d (body: %s)", libRec.Code, libRec.Body.String())
	}

	var lib domain.Library
	_ = json.NewDecoder(libRec.Body).Decode(&lib)

	// Execute synchronous library scan to ensure completion before assertions
	if _, err := catService.ScanLibrarySync(ctx, lib.ID); err != nil {
		t.Fatalf("ScanLibrarySync failed: %v", err)
	}

	// 3. Verify Extracted Metadata from Real Files
	tracks, total, err := catRepo.ListTracks(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list tracks failed: %v", err)
	}
	if total != 4 {
		t.Fatalf("expected 4 real tracks scanned, got %d", total)
	}

	var scannedFLAC *domain.Track
	for _, trk := range tracks {
		if trk.Title == "Wildlife Analysis" {
			scannedFLAC = trk
			break
		}
	}
	if scannedFLAC == nil {
		t.Fatal("Wildlife Analysis track was not discovered by real scanner")
	}

	if scannedFLAC.RawArtist != "Boards of Canada" || scannedFLAC.AlbumTitle != "Music Has the Right to Children" {
		t.Errorf("metadata mismatch: artist=%q, album=%q", scannedFLAC.RawArtist, scannedFLAC.AlbumTitle)
	}
	if scannedFLAC.Format != "flac" {
		t.Errorf("format mismatch: got %q, expected flac", scannedFLAC.Format)
	}

	// 4. Test User Annotations & Playlists on Real Track
	starBody := []byte(fmt.Sprintf(`{"itemType":"track","itemId":"%s","isStarred":true}`, scannedFLAC.ID))
	starReq := httptest.NewRequest(http.MethodPost, "/api/v1/annotations/star", bytes.NewReader(starBody))
	starReq.Header.Set("Authorization", "Bearer "+token)
	starRec := httptest.NewRecorder()
	router.ServeHTTP(starRec, starReq)
	if starRec.Code != http.StatusOK {
		t.Fatalf("star track failed: %d", starRec.Code)
	}

	pl := &domain.Playlist{
		ID:      "pl-real-1",
		Name:    "IDM Favorites",
		OwnerID: userID,
	}
	_ = plRepo.CreatePlaylist(ctx, pl)
	_ = plRepo.SetPlaylistTracks(ctx, pl.ID, userID, []string{scannedFLAC.ID})

	// 5. Test Move & Rename with 5-Tier PID Preservation
	newAlbumDir := filepath.Join(musicDir, "Boards of Canada", "MHTRTC - Relocated")
	_ = os.MkdirAll(newAlbumDir, 0755)
	renamedFLACPath := filepath.Join(newAlbumDir, "01 - Wildlife Analysis [Remastered].flac")
	if err := os.Rename(flacTrack, renamedFLACPath); err != nil {
		t.Fatalf("failed to move file: %v", err)
	}

	// Run synchronous scan again
	if _, err := catService.ScanLibrarySync(ctx, lib.ID); err != nil {
		t.Fatalf("rescan after rename failed: %v", err)
	}

	// Verify PID preservation: track moved, but PID and annotations persist
	favs, _ := annoRepo.GetFavoriteTracks(ctx, userID)
	if len(favs) != 1 {
		t.Fatalf("expected favorite track to be preserved after move/rename, got %d", len(favs))
	}
	if favs[0].PID != scannedFLAC.PID {
		t.Errorf("PID changed after move! old=%s, new=%s", scannedFLAC.PID, favs[0].PID)
	}

	// 6. Test Real Playable Audio Streaming with RFC 7233 Range Request
	streamReq := httptest.NewRequest(http.MethodGet, "/api/v1/stream/"+favs[0].ID, nil)
	streamReq.Header.Set("Authorization", "Bearer "+token)
	streamReq.Header.Set("Range", "bytes=0-100")
	streamRec := httptest.NewRecorder()
	router.ServeHTTP(streamRec, streamReq)

	if streamRec.Code != http.StatusPartialContent {
		t.Fatalf("expected 206 Partial Content, got %d", streamRec.Code)
	}
	if streamRec.Header().Get("Accept-Ranges") != "bytes" {
		t.Errorf("missing Accept-Ranges: bytes header")
	}
	if streamRec.Body.Len() != 101 {
		t.Errorf("expected 101 bytes, got %d", streamRec.Body.Len())
	}

	// 7. Test Real FFmpeg Transcoding Stream
	transcodeReq := httptest.NewRequest(http.MethodGet, "/api/v1/stream/"+favs[0].ID+"?format=mp3&bitrate=192", nil)
	transcodeReq.Header.Set("Authorization", "Bearer "+token)
	transcodeRec := httptest.NewRecorder()
	router.ServeHTTP(transcodeRec, transcodeReq)

	if transcodeRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for transcode stream, got %d", transcodeRec.Code)
	}
	if transcodeRec.Body.Len() < 500 {
		t.Errorf("expected real transcoded MP3 bytes, got only %d bytes", transcodeRec.Body.Len())
	}

	// 8. Test Live SQLite Online Backup & Integrity
	backupDest := filepath.Join(dir, "live_backup.db")
	if err := db.Backup(ctx, backupDest); err != nil {
		t.Fatalf("live sqlite backup failed: %v", err)
	}

	// Open backup database separately and verify integrity check & data presence
	backupLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	backupDB, err := repository.Open(backupDest, 5000, backupLogger)
	if err != nil {
		t.Fatalf("failed to open backup database: %v", err)
	}
	defer backupDB.Close()

	var integrity string
	_ = backupDB.Executor().QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity)
	if integrity != "ok" {
		t.Fatalf("backup database integrity check failed: %s", integrity)
	}

	var backupTrackCount int
	_ = backupDB.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&backupTrackCount)
	if backupTrackCount != 4 {
		t.Errorf("expected 4 tracks in backup database, found %d", backupTrackCount)
	}
}

func TestRealWorld_StorageOffline_Safety(t *testing.T) {
	srv, catService, catRepo, _, _, _, token, _, dir := setupRealServer(t)
	router := srv.Router()
	ctx := context.Background()

	// 1. Ingest real file
	musicDir := filepath.Join(dir, "temporary_mount")
	_ = os.MkdirAll(musicDir, 0755)
	flacTrack := filepath.Join(musicDir, "sample.flac")
	generateAudioFixture(t, flacTrack, "flac", "Sample Song", "Artist", "Album", 1, 2026, "")

	libBody, _ := json.Marshal(map[string]string{
		"name": "Removable Drive",
		"path": musicDir,
	})
	libReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/libraries", bytes.NewReader(libBody))
	libReq.Header.Set("Authorization", "Bearer "+token)
	libRec := httptest.NewRecorder()
	router.ServeHTTP(libRec, libReq)

	var lib domain.Library
	_ = json.NewDecoder(libRec.Body).Decode(&lib)

	// Run initial synchronous scan
	if _, err := catService.ScanLibrarySync(ctx, lib.ID); err != nil {
		t.Fatalf("initial scan failed: %v", err)
	}

	tracksBefore, totalBefore, _ := catRepo.ListTracks(ctx, 10, 0)
	if totalBefore != 1 {
		t.Fatalf("expected 1 track before unmount, got %d", totalBefore)
	}

	// 2. Simulate storage disconnect / offline mount by removing directory
	_ = os.RemoveAll(musicDir)

	// Scan attempt while storage is offline (should return error and abort)
	_, _ = catService.ScanLibrarySync(ctx, lib.ID)

	// 3. Verify safety guarantee: catalog is NOT wiped out when filesystem is inaccessible!
	tracksAfter, totalAfter, _ := catRepo.ListTracks(ctx, 10, 0)
	if totalAfter != 1 || len(tracksAfter) == 0 {
		t.Fatalf("CRITICAL BUG: tracks were mass-deleted when storage went offline! Total before=%d, after=%d", totalBefore, totalAfter)
	}

	// 4. Stream attempt on missing storage returns 404 / 500 error gracefully
	streamReq := httptest.NewRequest(http.MethodGet, "/api/v1/stream/"+tracksBefore[0].ID, nil)
	streamReq.Header.Set("Authorization", "Bearer "+token)
	streamRec := httptest.NewRecorder()
	router.ServeHTTP(streamRec, streamReq)

	if streamRec.Code != http.StatusNotFound && streamRec.Code != http.StatusInternalServerError {
		t.Errorf("expected error for offline storage stream, got code %d", streamRec.Code)
	}
}
