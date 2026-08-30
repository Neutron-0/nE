package repository

import (
	"context"
	"database/sql"
	"time"

	"ne/internal/domain"
	"github.com/google/uuid"
)

type AnnotationRepository struct {
	db *DB
}

func NewAnnotationRepository(db *DB) *AnnotationRepository {
	return &AnnotationRepository{db: db}
}

// ----------------- Star Operations -----------------

func (r *AnnotationRepository) SetTrackStar(ctx context.Context, userID, trackID string, starred bool) error {
	now := time.Now().UTC()
	query := `INSERT INTO user_track_annotations (user_id, track_id, is_starred, starred_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, track_id) DO UPDATE SET is_starred = excluded.is_starred, starred_at = excluded.starred_at`
	var starredAt *time.Time
	if starred {
		starredAt = &now
	}
	_, err := r.db.Executor().ExecContext(ctx, query, userID, trackID, starred, starredAt)
	return err
}

func (r *AnnotationRepository) SetAlbumStar(ctx context.Context, userID, albumID string, starred bool) error {
	now := time.Now().UTC()
	query := `INSERT INTO user_album_annotations (user_id, album_id, is_starred, starred_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, album_id) DO UPDATE SET is_starred = excluded.is_starred, starred_at = excluded.starred_at`
	var starredAt *time.Time
	if starred {
		starredAt = &now
	}
	_, err := r.db.Executor().ExecContext(ctx, query, userID, albumID, starred, starredAt)
	return err
}

func (r *AnnotationRepository) SetArtistStar(ctx context.Context, userID, artistID string, starred bool) error {
	now := time.Now().UTC()
	query := `INSERT INTO user_artist_annotations (user_id, artist_id, is_starred, starred_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, artist_id) DO UPDATE SET is_starred = excluded.is_starred, starred_at = excluded.starred_at`
	var starredAt *time.Time
	if starred {
		starredAt = &now
	}
	_, err := r.db.Executor().ExecContext(ctx, query, userID, artistID, starred, starredAt)
	return err
}

// ----------------- Rating Operations -----------------

func (r *AnnotationRepository) SetTrackRating(ctx context.Context, userID, trackID string, rating int) error {
	if rating < 0 || rating > 5 {
		return domain.ErrInvalidInput("Rating must be between 0 and 5")
	}
	now := time.Now().UTC()
	query := `INSERT INTO user_track_annotations (user_id, track_id, rating, rated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, track_id) DO UPDATE SET rating = excluded.rating, rated_at = excluded.rated_at`
	var ratedAt *time.Time
	if rating > 0 {
		ratedAt = &now
	}
	_, err := r.db.Executor().ExecContext(ctx, query, userID, trackID, rating, ratedAt)
	return err
}

func (r *AnnotationRepository) SetAlbumRating(ctx context.Context, userID, albumID string, rating int) error {
	if rating < 0 || rating > 5 {
		return domain.ErrInvalidInput("Rating must be between 0 and 5")
	}
	now := time.Now().UTC()
	query := `INSERT INTO user_album_annotations (user_id, album_id, rating, rated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, album_id) DO UPDATE SET rating = excluded.rating, rated_at = excluded.rated_at`
	var ratedAt *time.Time
	if rating > 0 {
		ratedAt = &now
	}
	_, err := r.db.Executor().ExecContext(ctx, query, userID, albumID, rating, ratedAt)
	return err
}

// ----------------- Playback & Scrobble Operations -----------------

func (r *AnnotationRepository) RecordPlayback(ctx context.Context, userID, trackID, playerName string, durationPlayed float64, completed bool) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		now := time.Now().UTC()
		histID := uuid.NewString()

		// Insert into history log
		histQuery := `INSERT INTO playback_history (id, user_id, track_id, player_name, played_at, duration_played, completed)
			VALUES (?, ?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(ctx, histQuery, histID, userID, trackID, playerName, now, durationPlayed, completed); err != nil {
			return err
		}

		// Update track annotation play count if completed
		if completed {
			annoQuery := `INSERT INTO user_track_annotations (user_id, track_id, play_count, last_played_at)
				VALUES (?, ?, 1, ?)
				ON CONFLICT(user_id, track_id) DO UPDATE SET play_count = play_count + 1, last_played_at = excluded.last_played_at`
			if _, err := tx.ExecContext(ctx, annoQuery, userID, trackID, now); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AnnotationRepository) GetRecentHistory(ctx context.Context, userID string, limit int) ([]*domain.PlaybackRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT h.id, h.user_id, h.track_id, h.player_name, h.played_at, h.duration_played, h.completed,
		t.title, t.raw_artist, al.title
		FROM playback_history h
		JOIN tracks t ON h.track_id = t.id
		JOIN albums al ON t.album_id = al.id
		WHERE h.user_id = ?
		ORDER BY h.played_at DESC
		LIMIT ?`
	rows, err := r.db.Executor().QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.PlaybackRecord
	for rows.Next() {
		var rec domain.PlaybackRecord
		var pName sql.NullString
		var trackTitle, artistName, albumTitle string
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.TrackID, &pName, &rec.PlayedAt, &rec.DurationPlayed, &rec.Completed,
			&trackTitle, &artistName, &albumTitle); err != nil {
			return nil, err
		}
		if pName.Valid {
			rec.PlayerName = pName.String
		}
		records = append(records, &rec)
	}
	return records, nil
}

func (r *AnnotationRepository) GetFavoriteTracks(ctx context.Context, userID string) ([]*domain.Track, error) {
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title,
		t.raw_artist, t.album_id, al.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle,
		t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth, t.channels, t.format, t.codec,
		t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM user_track_annotations a
		JOIN tracks t ON a.track_id = t.id
		JOIN albums al ON t.album_id = al.id
		LEFT JOIN artists ar ON al.album_artist_id = ar.id
		WHERE a.user_id = ? AND a.is_starred = 1
		ORDER BY a.starred_at DESC`
	rows, err := r.db.Executor().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*domain.Track
	for rows.Next() {
		var t domain.Track
		var discSub, mbz sql.NullString
		if err := rows.Scan(
			&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle,
			&t.RawArtist, &t.AlbumID, &t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub,
			&t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth, &t.Channels, &t.Format, &t.Codec,
			&t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
			&t.HasEmbeddedCover, &mbz, &t.MTime, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if discSub.Valid {
			t.DiscSubtitle = discSub.String
		}
		if mbz.Valid {
			t.MbzTrackID = mbz.String
		}
		t.IsStarred = true
		tracks = append(tracks, &t)
	}
	return tracks, nil
}
