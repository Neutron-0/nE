package service_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"ne/internal/domain"
	"ne/internal/media"
	"ne/internal/repository"
	"ne/internal/service"
)

func setupTestServices(t *testing.T) (
	*service.AnnotationService,
	*service.PlaylistService,
	*service.ArtworkService,
	*service.StreamService,
	*repository.CatalogRepository,
	*repository.UserRepository,
	*repository.DB,
	string,
) {
	t.Helper()
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := repository.Open(filepath.Join(dir, "service_test.db"), 5000, logger)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	plRepo := repository.NewPlaylistRepository(db)

	cacheDir := filepath.Join(dir, "cache")
	_ = os.MkdirAll(cacheDir, 0750)

	annoService := service.NewAnnotationService(annoRepo)
	plService := service.NewPlaylistService(plRepo)
	artService := service.NewArtworkService(catRepo, cacheDir)
	streamService := service.NewStreamService(catRepo, media.NewTranscoder(2))

	return annoService, plService, artService, streamService, catRepo, userRepo, db, dir
}

func TestAnnotationAndPlaylistServices(t *testing.T) {
	annoService, plService, _, _, catRepo, userRepo, db, _ := setupTestServices(t)
	ctx := context.Background()

	// 1. Setup User and Library
	user := &domain.User{ID: "user-svc", Username: "audiophile", PasswordHash: "hash"}
	_ = userRepo.Create(ctx, user)

	lib := &domain.Library{ID: "lib-svc", Name: "Lib", Path: "/music"}
	_ = catRepo.CreateLibrary(ctx, lib)

	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Boards of Canada", "Boards of Canada", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "Music Has the Right to Children", SortTitle: "Music Has the Right to Children", AlbumArtistID: art.ID})
	trk := &domain.Track{ID: "trk-telephasic", PID: "pid-telephasic", LibraryID: lib.ID, Path: "/music/telephasic.flac", FolderPath: "/music", Filename: "telephasic.flac", Title: "Telephasic Workshop", SortTitle: "Telephasic Workshop", RawArtist: "Boards of Canada", AlbumID: alb.ID, Duration: 395.0, TrackNumber: 5, DiscNumber: 1, MTime: 1000}
	_ = catRepo.UpsertTrack(ctx, db.Executor(), trk)

	// 2. Star Track and Check Favorites
	if err := annoService.StarItem(ctx, user.ID, "track", trk.ID, true); err != nil {
		t.Fatalf("star item failed: %v", err)
	}

	favs, err := annoService.GetFavorites(ctx, user.ID)
	if err != nil {
		t.Fatalf("get favorites failed: %v", err)
	}
	if len(favs) != 1 || favs[0].ID != trk.ID {
		t.Errorf("expected track in favorites, got %v", favs)
	}

	// 3. Scrobble Playback
	if err := annoService.Scrobble(ctx, user.ID, trk.ID, "Web Player", 395.0, true); err != nil {
		t.Fatalf("scrobble failed: %v", err)
	}

	history, err := annoService.GetRecentHistory(ctx, user.ID, 5)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history) != 1 || history[0].TrackID != trk.ID {
		t.Errorf("expected scrobble in history, got %v", history)
	}

	// 4. Playlist Operations via Service
	pl, err := plService.CreatePlaylist(ctx, user.ID, "Downtempo Classics", "90s vibes", false)
	if err != nil {
		t.Fatalf("create playlist failed: %v", err)
	}

	// Set tracks with correct (ctx, userID, playlistID, trackIDs)
	if err := plService.SetPlaylistTracks(ctx, user.ID, pl.ID, []string{trk.ID}); err != nil {
		t.Fatalf("set playlist tracks failed: %v", err)
	}

	fetchedPl, err := plService.GetPlaylist(ctx, pl.ID)
	if err != nil {
		t.Fatalf("get playlist failed: %v", err)
	}
	if len(fetchedPl.Tracks) != 1 || fetchedPl.Tracks[0].TrackID != trk.ID {
		t.Errorf("unexpected tracks in playlist: %+v", fetchedPl.Tracks)
	}
}

func TestStreamService_SecurityJail(t *testing.T) {
	_, _, _, streamService, catRepo, _, db, dir := setupTestServices(t)
	ctx := context.Background()

	// Library is inside /music
	musicDir := filepath.Join(dir, "music")
	_ = os.MkdirAll(musicDir, 0755)

	lib := &domain.Library{ID: "lib-sec", Name: "Lib", Path: musicDir}
	_ = catRepo.CreateLibrary(ctx, lib)

	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Artist", "Artist", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "Album", SortTitle: "Album", AlbumArtistID: art.ID})

	// Malicious track pointing OUTSIDE library root (e.g. into root directory)
	outsideFile := filepath.Join(dir, "secrets.txt")
	_ = os.WriteFile(outsideFile, []byte("SECRET_DATA"), 0644)

	maliciousTrack := &domain.Track{
		ID:         "trk-malicious",
		PID:        "pid-malicious",
		LibraryID:  lib.ID,
		Path:       outsideFile, // outside musicDir!
		FolderPath: dir,
		Filename:   "secrets.txt",
		Title:      "Secrets",
		SortTitle:  "Secrets",
		RawArtist:  "Hacker",
		AlbumID:    alb.ID,
		Duration:   100.0,
		Format:     "flac",
	}
	_ = catRepo.UpsertTrack(ctx, db.Executor(), maliciousTrack)

	// Stream attempt should be rejected with 403 Forbidden
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stream/trk-malicious", nil)
	rec := httptest.NewRecorder()

	err := streamService.StreamTrack(ctx, rec, req, maliciousTrack.ID)
	if err == nil {
		t.Fatal("expected Forbidden error for file outside library root, got nil")
	}

	appErr, ok := err.(*domain.AppError)
	if !ok || appErr.HTTPStatus != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden error, got: %v", err)
	}
}
