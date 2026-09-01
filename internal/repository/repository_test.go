package repository_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"ne/internal/domain"
	"ne/internal/repository"
)

func setupTestRepositories(t *testing.T) (*repository.DB, *repository.CatalogRepository, *repository.AnnotationRepository, *repository.PlaylistRepository) {
	t.Helper()
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := repository.Open(filepath.Join(dir, "repo_test.db"), 5000, logger)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	catRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	plRepo := repository.NewPlaylistRepository(db)

	return db, catRepo, annoRepo, plRepo
}

func TestPlaylistOperations(t *testing.T) {
	db, catRepo, _, plRepo := setupTestRepositories(t)
	ctx := context.Background()

	// Setup dummy user, library, album, tracks
	userRepo := repository.NewUserRepository(db)
	user := &domain.User{ID: "user-1", Username: "listener", PasswordHash: "hash"}
	_ = userRepo.Create(ctx, user)

	lib := &domain.Library{ID: "lib-1", Name: "Lib", Path: "/music"}
	_ = catRepo.CreateLibrary(ctx, lib)

	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Daft Punk", "Daft Punk", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "Discovery", SortTitle: "Discovery", AlbumArtistID: art.ID})

	t1 := &domain.Track{ID: "t1", PID: "p1", LibraryID: lib.ID, Path: "/music/01.flac", FolderPath: "/music", Filename: "01.flac", Title: "One More Time", SortTitle: "One More Time", RawArtist: "Daft Punk", AlbumID: alb.ID, Duration: 320.0, TrackNumber: 1, DiscNumber: 1, MTime: 1000}
	t2 := &domain.Track{ID: "t2", PID: "p2", LibraryID: lib.ID, Path: "/music/02.flac", FolderPath: "/music", Filename: "02.flac", Title: "Aerodynamic", SortTitle: "Aerodynamic", RawArtist: "Daft Punk", AlbumID: alb.ID, Duration: 212.0, TrackNumber: 2, DiscNumber: 1, MTime: 1000}
	_ = catRepo.UpsertTrack(ctx, db.Executor(), t1)
	_ = catRepo.UpsertTrack(ctx, db.Executor(), t2)

	// 1. Create Playlist
	pl := &domain.Playlist{
		ID:      "pl-1",
		Name:    "Electronic Gems",
		Comment: "Best electronic tracks",
		OwnerID: user.ID,
	}
	if err := plRepo.CreatePlaylist(ctx, pl); err != nil {
		t.Fatalf("create playlist failed: %v", err)
	}

	// 2. Set Tracks in Playlist
	if err := plRepo.SetPlaylistTracks(ctx, pl.ID, user.ID, []string{t1.ID, t2.ID}); err != nil {
		t.Fatalf("set playlist tracks failed: %v", err)
	}

	// 3. Get Playlist Tracks
	tracks, err := plRepo.GetPlaylistTracks(ctx, pl.ID)
	if err != nil {
		t.Fatalf("get playlist tracks failed: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks loaded, got %d", len(tracks))
	}
	if tracks[0].TrackID != t1.ID || tracks[1].TrackID != t2.ID {
		t.Errorf("unexpected playlist track order")
	}

	// 4. Reorder Tracks (swap position)
	if err := plRepo.SetPlaylistTracks(ctx, pl.ID, user.ID, []string{t2.ID, t1.ID}); err != nil {
		t.Fatalf("reorder playlist failed: %v", err)
	}

	reorderedTracks, _ := plRepo.GetPlaylistTracks(ctx, pl.ID)
	if reorderedTracks[0].TrackID != t2.ID || reorderedTracks[1].TrackID != t1.ID {
		t.Errorf("tracks were not reordered properly")
	}

	// 5. Delete Playlist
	if err := plRepo.DeletePlaylist(ctx, pl.ID, user.ID); err != nil {
		t.Fatalf("delete playlist failed: %v", err)
	}

	_, err = plRepo.GetPlaylistByID(ctx, pl.ID)
	if err == nil {
		t.Error("expected error getting deleted playlist, got nil")
	}
}

func TestAnnotationsAndHistory(t *testing.T) {
	db, catRepo, annoRepo, _ := setupTestRepositories(t)
	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	user := &domain.User{ID: "user-anno", Username: "musiclover", PasswordHash: "hash"}
	_ = userRepo.Create(ctx, user)

	lib := &domain.Library{ID: "lib-anno", Name: "Lib", Path: "/music"}
	_ = catRepo.CreateLibrary(ctx, lib)

	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Aphex Twin", "Aphex Twin", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "Selected Ambient Works", SortTitle: "Selected Ambient Works", AlbumArtistID: art.ID})
	trk := &domain.Track{ID: "trk-xtal", PID: "pid-xtal", LibraryID: lib.ID, Path: "/music/xtal.flac", FolderPath: "/music", Filename: "xtal.flac", Title: "Xtal", SortTitle: "Xtal", RawArtist: "Aphex Twin", AlbumID: alb.ID, Duration: 294.0, TrackNumber: 1, DiscNumber: 1, MTime: 1000}
	_ = catRepo.UpsertTrack(ctx, db.Executor(), trk)

	// 1. Star Track
	if err := annoRepo.SetTrackStar(ctx, user.ID, trk.ID, true); err != nil {
		t.Fatalf("star track failed: %v", err)
	}

	starred, err := annoRepo.GetFavoriteTracks(ctx, user.ID)
	if err != nil {
		t.Fatalf("get favorite tracks failed: %v", err)
	}
	if len(starred) != 1 || starred[0].ID != trk.ID {
		t.Errorf("expected 1 starred track (%s), got: %v", trk.ID, starred)
	}

	// 2. Rate Album
	if err := annoRepo.SetAlbumRating(ctx, user.ID, alb.ID, 5); err != nil {
		t.Fatalf("rate album failed: %v", err)
	}

	// 3. Scrobble & Playback History
	if err := annoRepo.RecordPlayback(ctx, user.ID, trk.ID, "Web Player", 294.0, true); err != nil {
		t.Fatalf("record playback failed: %v", err)
	}

	history, err := annoRepo.GetRecentHistory(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("list history failed: %v", err)
	}
	if len(history) != 1 || history[0].TrackID != trk.ID || !history[0].Completed {
		t.Errorf("unexpected playback history: %+v", history)
	}

	// 4. Unstar Track
	if err := annoRepo.SetTrackStar(ctx, user.ID, trk.ID, false); err != nil {
		t.Fatalf("unstar track failed: %v", err)
	}
	starredAfter, _ := annoRepo.GetFavoriteTracks(ctx, user.ID)
	if len(starredAfter) != 0 {
		t.Errorf("expected 0 starred tracks after unstar, got %d", len(starredAfter))
	}
}

func TestFullTextSearchFTS5(t *testing.T) {
	db, catRepo, _, _ := setupTestRepositories(t)
	ctx := context.Background()

	lib := &domain.Library{ID: "lib-fts", Name: "Lib", Path: "/music"}
	_ = catRepo.CreateLibrary(ctx, lib)

	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Björk", "Bjork", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "Homogenic", SortTitle: "Homogenic", AlbumArtistID: art.ID})
	trk := &domain.Track{ID: "trk-joga", PID: "pid-joga", LibraryID: lib.ID, Path: "/music/joga.flac", FolderPath: "/music", Filename: "joga.flac", Title: "Jóga", SortTitle: "Joga", RawArtist: "Björk", AlbumID: alb.ID, Duration: 305.0, TrackNumber: 3, DiscNumber: 1, MTime: 1000}
	_ = catRepo.UpsertTrack(ctx, db.Executor(), trk)
	_ = catRepo.IndexTrackFTS(ctx, db.Executor(), trk.ID, "Jóga", "Björk", "Homogenic", "Electronic")

	// Search with diacritics
	res1, err := catRepo.Search(ctx, "Jóga", 10)
	if err != nil {
		t.Fatalf("search Jóga failed: %v", err)
	}
	if len(res1.Tracks) == 0 {
		t.Error("expected at least 1 track result searching for 'Jóga'")
	}

	// Search stripped diacritics
	res2, err := catRepo.Search(ctx, "Joga", 10)
	if err != nil {
		t.Fatalf("search Joga failed: %v", err)
	}
	if len(res2.Tracks) == 0 {
		t.Error("expected at least 1 track result searching for stripped 'Joga'")
	}

	// Search Artist
	res3, err := catRepo.Search(ctx, "Bjork", 10)
	if err != nil {
		t.Fatalf("search Bjork failed: %v", err)
	}
	if len(res3.Artists) == 0 && len(res3.Tracks) == 0 {
		t.Error("expected search results for artist 'Bjork'")
	}

	// Verify FTS Delete Trigger
	_ = catRepo.DeleteTrackByPath(ctx, db.Executor(), trk.Path)
	resAfterDelete, _ := catRepo.Search(ctx, "Joga", 10)
	if len(resAfterDelete.Tracks) != 0 {
		t.Errorf("expected 0 tracks in FTS after track deletion, got %d", len(resAfterDelete.Tracks))
	}
}

func TestRecentlyPlayedDeduplication(t *testing.T) {
	db, catRepo, annoRepo, _ := setupTestRepositories(t)
	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	user := &domain.User{ID: "user-recent", Username: "recent_listener", PasswordHash: "hash"}
	_ = userRepo.Create(ctx, user)

	lib := &domain.Library{ID: "lib-recent", Name: "Lib", Path: "/music"}
	_ = catRepo.CreateLibrary(ctx, lib)

	art, _ := catRepo.FindOrCreateArtist(ctx, db.Executor(), "Artist 1", "Artist 1", "")
	alb, _ := catRepo.FindOrCreateAlbum(ctx, db.Executor(), &domain.Album{Title: "Album 1", SortTitle: "Album 1", AlbumArtistID: art.ID})
	trk := &domain.Track{ID: "trk-repeat", PID: "pid-repeat", LibraryID: lib.ID, Path: "/music/repeat.flac", FolderPath: "/music", Filename: "repeat.flac", Title: "Repeated Song", SortTitle: "Repeated Song", RawArtist: "Artist 1", AlbumID: alb.ID, Duration: 200.0, TrackNumber: 1, DiscNumber: 1, MTime: 1000}
	_ = catRepo.UpsertTrack(ctx, db.Executor(), trk)

	// Scrobble the exact same track 3 times
	for i := 0; i < 3; i++ {
		_ = annoRepo.RecordPlayback(ctx, user.ID, trk.ID, "web", 200.0, true)
	}

	recent, err := catRepo.GetRecentlyPlayedTracks(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("get recently played failed: %v", err)
	}

	if len(recent) != 1 {
		t.Errorf("expected exactly 1 deduplicated track in recently played, got %d", len(recent))
	}
}
