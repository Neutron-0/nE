package scanner_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"ne/internal/config"
	"ne/internal/domain"
	"ne/internal/repository"
	"ne/internal/scanner"
)

func setupTestScanner(t *testing.T) (*scanner.Scanner, *repository.CatalogRepository, *repository.DB, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "scanner_test.db")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := repository.Open(dbPath, 5000, logger)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	catalogRepo := repository.NewCatalogRepository(db)
	cfg := &config.ScannerConfig{
		WorkerCount:  2,
		BatchSize:    10,
		CleanEnglish: true,
	}

	sc := scanner.NewScanner(db, catalogRepo, cfg, logger)
	return sc, catalogRepo, db, dir
}

func TestScanner_EndToEnd(t *testing.T) {
	sc, catalogRepo, _, dir := setupTestScanner(t)
	ctx := context.Background()

	// 1. Create a synthetic music library directory
	musicDir := filepath.Join(dir, "music")
	albumDir := filepath.Join(musicDir, "The Beatles", "Abbey Road")
	_ = os.MkdirAll(albumDir, 0755)

	// Create test audio files
	track1Path := filepath.Join(albumDir, "01 - Come Together.mp3")
	track2Path := filepath.Join(albumDir, "02 - Something.flac")
	_ = os.WriteFile(track1Path, []byte("fake mp3 audio data 1234567890"), 0644)
	_ = os.WriteFile(track2Path, []byte("fake flac audio data 1234567890"), 0644)

	// Create Library in DB
	lib := &domain.Library{
		ID:   "lib-1",
		Name: "Main Collection",
		Path: musicDir,
	}
	if err := catalogRepo.CreateLibrary(ctx, lib); err != nil {
		t.Fatalf("failed to create library: %v", err)
	}

	// 2. Run Scan
	progress, err := sc.ScanLibrary(ctx, "lib-1")
	if err != nil {
		t.Fatalf("scan library failed: %v", err)
	}

	if progress.FilesFound != 2 {
		t.Errorf("expected 2 files found, got %d", progress.FilesFound)
	}
	if progress.FilesProcessed != 2 {
		t.Errorf("expected 2 files processed, got %d", progress.FilesProcessed)
	}

	// 3. Verify Database Records
	artists, _, err := catalogRepo.ListArtists(ctx, 10, 0)
	if err != nil || len(artists) == 0 {
		t.Fatalf("expected artists in catalog, got error: %v, count: %d", err, len(artists))
	}
	if artists[0].Name != "The Beatles" && artists[0].Name != "Abbey Road" {
		t.Logf("Artist name extracted from path: %q", artists[0].Name)
	}

	albums, _, err := catalogRepo.ListAlbums(ctx, 10, 0)
	if err != nil || len(albums) == 0 {
		t.Fatalf("expected albums in catalog, got error: %v, count: %d", err, len(albums))
	}

	tracks, totalTracks, err := catalogRepo.ListTracks(ctx, 10, 0)
	if err != nil || totalTracks != 2 {
		t.Fatalf("expected 2 tracks, got %d (err: %v)", totalTracks, err)
	}
	if tracks[0].Title == "" {
		t.Error("expected non-empty track title")
	}

	// 4. Verify Idempotency on second scan
	progress2, err := sc.ScanLibrary(ctx, "lib-1")
	if err != nil {
		t.Fatalf("second scan failed: %v", err)
	}
	if progress2.FilesProcessed != 2 {
		t.Errorf("expected 2 files processed in second scan, got %d", progress2.FilesProcessed)
	}

	// 5. Test Cleanup: Delete track 2 and re-scan
	_ = os.Remove(track2Path)
	progress3, err := sc.ScanLibrary(ctx, "lib-1")
	if err != nil {
		t.Fatalf("third scan failed: %v", err)
	}
	if progress3.FilesFound != 1 {
		t.Errorf("expected 1 file found after delete, got %d", progress3.FilesFound)
	}

	_, remainingTracks, _ := catalogRepo.ListTracks(ctx, 10, 0)
	if remainingTracks != 1 {
		t.Errorf("expected 1 track remaining after cleanup, got %d", remainingTracks)
	}
}
