package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ne/internal/domain"
	"github.com/google/uuid"
)

type PlaylistRepository struct {
	db *DB
}

func NewPlaylistRepository(db *DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) CreatePlaylist(ctx context.Context, p *domain.Playlist) error {
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	query := `INSERT INTO playlists (id, name, comment, owner_id, is_public, duration, track_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Executor().ExecContext(ctx, query,
		p.ID, p.Name, p.Comment, p.OwnerID, p.IsPublic, p.Duration, p.TrackCount, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PlaylistRepository) GetPlaylistByID(ctx context.Context, id string) (*domain.Playlist, error) {
	query := `SELECT id, name, comment, owner_id, is_public, duration, track_count, created_at, updated_at
		FROM playlists WHERE id = ?`
	var p domain.Playlist
	var comment sql.NullString
	err := r.db.Executor().QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &comment, &p.OwnerID, &p.IsPublic, &p.Duration, &p.TrackCount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("Playlist", id)
		}
		return nil, err
	}
	if comment.Valid {
		p.Comment = comment.String
	}
	return &p, nil
}

func (r *PlaylistRepository) ListPlaylistsForUser(ctx context.Context, userID string) ([]*domain.Playlist, error) {
	query := `SELECT id, name, comment, owner_id, is_public, duration, track_count, created_at, updated_at
		FROM playlists WHERE owner_id = ? OR is_public = 1 ORDER BY name COLLATE NOCASE ASC`
	rows, err := r.db.Executor().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*domain.Playlist
	for rows.Next() {
		var p domain.Playlist
		var comment sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &comment, &p.OwnerID, &p.IsPublic, &p.Duration, &p.TrackCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if comment.Valid {
			p.Comment = comment.String
		}
		playlists = append(playlists, &p)
	}
	return playlists, nil
}

func (r *PlaylistRepository) UpdatePlaylist(ctx context.Context, p *domain.Playlist) error {
	p.UpdatedAt = time.Now().UTC()
	query := `UPDATE playlists SET name = ?, comment = ?, is_public = ?, updated_at = ? WHERE id = ? AND owner_id = ?`
	res, err := r.db.Executor().ExecContext(ctx, query, p.Name, p.Comment, p.IsPublic, p.UpdatedAt, p.ID, p.OwnerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound("Playlist", p.ID)
	}
	return nil
}

func (r *PlaylistRepository) DeletePlaylist(ctx context.Context, id, userID string) error {
	query := `DELETE FROM playlists WHERE id = ? AND owner_id = ?`
	res, err := r.db.Executor().ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound("Playlist", id)
	}
	return nil
}

func (r *PlaylistRepository) SetPlaylistTracks(ctx context.Context, playlistID, ownerID string, trackIDs []string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// Verify ownership
		var exists bool
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM playlists WHERE id = ? AND owner_id = ?`, playlistID, ownerID).Scan(&exists)
		if err != nil {
			return domain.ErrForbidden("Cannot modify playlist owned by another user")
		}

		// Delete existing tracks
		_, err = tx.ExecContext(ctx, `DELETE FROM playlist_tracks WHERE playlist_id = ?`, playlistID)
		if err != nil {
			return err
		}

		// Insert new tracks in order
		var totalDuration float64
		now := time.Now().UTC()
		for pos, trkID := range trackIDs {
			ptID := uuid.NewString()
			_, err = tx.ExecContext(ctx,
				`INSERT INTO playlist_tracks (id, playlist_id, track_id, position, added_at) VALUES (?, ?, ?, ?, ?)`,
				ptID, playlistID, trkID, pos, now,
			)
			if err != nil {
				return err
			}
			var dur float64
			_ = tx.QueryRowContext(ctx, `SELECT duration FROM tracks WHERE id = ?`, trkID).Scan(&dur)
			totalDuration += dur
		}

		// Update playlist track count and duration
		_, err = tx.ExecContext(ctx,
			`UPDATE playlists SET track_count = ?, duration = ?, updated_at = ? WHERE id = ?`,
			len(trackIDs), totalDuration, now, playlistID,
		)
		return err
	})
}

func (r *PlaylistRepository) GetPlaylistTracks(ctx context.Context, playlistID string) ([]*domain.PlaylistTrack, error) {
	query := `SELECT pt.id, pt.playlist_id, pt.track_id, pt.position, pt.added_at,
		t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title,
		t.raw_artist, t.album_id, al.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle,
		t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth, t.channels, t.format, t.codec,
		t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM playlist_tracks pt
		JOIN tracks t ON pt.track_id = t.id
		JOIN albums al ON t.album_id = al.id
		LEFT JOIN artists ar ON al.album_artist_id = ar.id
		WHERE pt.playlist_id = ?
		ORDER BY pt.position ASC`
	rows, err := r.db.Executor().QueryContext(ctx, query, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pts []*domain.PlaylistTrack
	for rows.Next() {
		var pt domain.PlaylistTrack
		var t domain.Track
		var discSub, mbz sql.NullString
		if err := rows.Scan(
			&pt.ID, &pt.PlaylistID, &pt.TrackID, &pt.Position, &pt.AddedAt,
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
		pt.Track = &t
		pts = append(pts, &pt)
	}
	return pts, nil
}
