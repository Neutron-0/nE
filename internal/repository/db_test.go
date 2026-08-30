package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"ne/internal/repository"
)

func newTestDB(t *testing.T) *repository.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_ne.db")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := repository.Open(dbPath, 5000, logger)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDatabaseMigration(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Verify schema_migrations has version 1
	var count int
	err := db.Executor().QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected version 1 applied, got count %d", count)
	}

	// Verify tables exist
	tables := []string{
		"libraries", "users", "refresh_tokens", "artists", "albums",
		"tracks", "track_artists", "genres", "track_genres", "album_genres",
		"user_track_annotations", "user_album_annotations", "user_artist_annotations",
		"playlists", "playlist_tracks", "playback_history", "scan_runs",
		"track_search_fts",
	}

	for _, table := range tables {
		var name string
		err := db.Executor().QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type IN ('table', 'shadow') AND name = ?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q does not exist: %v", table, err)
		}
	}
}

func TestForeignKeyCascades(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Insert Library, Artist, Album, Track
	_, err := db.Executor().ExecContext(ctx, `INSERT INTO libraries (id, name, path) VALUES ('lib-1', 'Main', '/music')`)
	if err != nil {
		t.Fatalf("insert library failed: %v", err)
	}

	_, err = db.Executor().ExecContext(ctx, `INSERT INTO artists (id, name, sort_name) VALUES ('art-1', 'Radiohead', 'Radiohead')`)
	if err != nil {
		t.Fatalf("insert artist failed: %v", err)
	}

	_, err = db.Executor().ExecContext(ctx, `INSERT INTO albums (id, title, sort_title, album_artist_id, year) VALUES ('alb-1', 'OK Computer', 'OK Computer', 'art-1', 1997)`)
	if err != nil {
		t.Fatalf("insert album failed: %v", err)
	}

	_, err = db.Executor().ExecContext(ctx, `INSERT INTO tracks (id, pid, library_id, path, folder_path, filename, title, sort_title, raw_artist, album_id, duration, format, codec, mtime)
		VALUES ('trk-1', 'pid-1', 'lib-1', '/music/airbag.flac', '/music', 'airbag.flac', 'Airbag', 'Airbag', 'Radiohead', 'alb-1', 284.0, 'flac', 'flac', 1000)`)
	if err != nil {
		t.Fatalf("insert track failed: %v", err)
	}

	// Delete Library -> Should cascade to delete track
	_, err = db.Executor().ExecContext(ctx, `DELETE FROM libraries WHERE id = 'lib-1'`)
	if err != nil {
		t.Fatalf("delete library failed: %v", err)
	}

	var trackCount int
	err = db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks WHERE id = 'trk-1'`).Scan(&trackCount)
	if err != nil {
		t.Fatalf("query track failed: %v", err)
	}
	if trackCount != 0 {
		t.Errorf("expected track to be cascade deleted, count = %d", trackCount)
	}
}

func TestTransactionRollback(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	err := db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO artists (id, name, sort_name) VALUES ('art-rollback', 'Temporary', 'Temporary')`)
		if err != nil {
			return err
		}
		return errors.New("deliberate failure")
	})

	if err == nil {
		t.Fatal("expected transaction error")
	}

	var count int
	_ = db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM artists WHERE id = 'art-rollback'`).Scan(&count)
	if count != 0 {
		t.Errorf("expected rolled back artist to not exist, count = %d", count)
	}
}
