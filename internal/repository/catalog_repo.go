package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"ne/internal/domain"
	"github.com/google/uuid"
)

type CatalogRepository struct {
	db *DB
}

func NewCatalogRepository(db *DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

// ----------------- Library Operations -----------------

func (r *CatalogRepository) CreateLibrary(ctx context.Context, lib *domain.Library) error {
	now := time.Now().UTC()
	lib.CreatedAt = now
	lib.UpdatedAt = now
	query := `INSERT INTO libraries (id, name, path, scan_status, track_count, album_count, total_bytes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Executor().ExecContext(ctx, query,
		lib.ID, lib.Name, lib.Path, lib.ScanStatus, lib.TrackCount, lib.AlbumCount, lib.TotalBytes, lib.CreatedAt, lib.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) GetLibraryByID(ctx context.Context, id string) (*domain.Library, error) {
	query := `SELECT id, name, path, last_scanned_at, scan_status, track_count, album_count, total_bytes, created_at, updated_at
		FROM libraries WHERE id = ?`
	var lib domain.Library
	var lastScanned sql.NullTime
	err := r.db.Executor().QueryRowContext(ctx, query, id).Scan(
		&lib.ID, &lib.Name, &lib.Path, &lastScanned, &lib.ScanStatus, &lib.TrackCount, &lib.AlbumCount, &lib.TotalBytes, &lib.CreatedAt, &lib.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("Library", id)
		}
		return nil, err
	}
	if lastScanned.Valid {
		lib.LastScannedAt = &lastScanned.Time
	}
	return &lib, nil
}

func (r *CatalogRepository) ListLibraries(ctx context.Context) ([]*domain.Library, error) {
	query := `SELECT id, name, path, last_scanned_at, scan_status, track_count, album_count, total_bytes, created_at, updated_at
		FROM libraries ORDER BY name ASC`
	rows, err := r.db.Executor().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libs []*domain.Library
	for rows.Next() {
		var lib domain.Library
		var lastScanned sql.NullTime
		if err := rows.Scan(
			&lib.ID, &lib.Name, &lib.Path, &lastScanned, &lib.ScanStatus, &lib.TrackCount, &lib.AlbumCount, &lib.TotalBytes, &lib.CreatedAt, &lib.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if lastScanned.Valid {
			lib.LastScannedAt = &lastScanned.Time
		}
		libs = append(libs, &lib)
	}
	return libs, nil
}

func (r *CatalogRepository) UpdateLibraryScanStatus(ctx context.Context, id, status string) error {
	query := `UPDATE libraries SET scan_status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Executor().ExecContext(ctx, query, status, time.Now().UTC(), id)
	return err
}

func (r *CatalogRepository) UpdateLibraryStats(ctx context.Context, id string, tracks, albums int, bytes int64) error {
	now := time.Now().UTC()
	query := `UPDATE libraries SET
		scan_status = 'idle',
		last_scanned_at = ?,
		track_count = ?,
		album_count = ?,
		total_bytes = ?,
		updated_at = ?
		WHERE id = ?`
	_, err := r.db.Executor().ExecContext(ctx, query, now, tracks, albums, bytes, now, id)
	return err
}

// ----------------- Artist Operations -----------------

func (r *CatalogRepository) FindOrCreateArtist(ctx context.Context, exec DBExecutor, name, sortName, mbzID string) (*domain.Artist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Unknown Artist"
	}
	if sortName == "" {
		sortName = name
	}

	var art domain.Artist
	var mbz sql.NullString
	query := `SELECT id, name, sort_name, mbz_artist_id, album_count, track_count, created_at, updated_at
		FROM artists WHERE name = ? COLLATE NOCASE`
	err := exec.QueryRowContext(ctx, query, name).Scan(
		&art.ID, &art.Name, &art.SortName, &mbz, &art.AlbumCount, &art.TrackCount, &art.CreatedAt, &art.UpdatedAt,
	)
	if err == nil {
		if mbz.Valid {
			art.MbzArtistID = mbz.String
		}
		return &art, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Create new artist
	art = domain.Artist{
		ID:          uuid.NewString(),
		Name:        name,
		SortName:    sortName,
		MbzArtistID: mbzID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	insertQuery := `INSERT INTO artists (id, name, sort_name, mbz_artist_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = exec.ExecContext(ctx, insertQuery, art.ID, art.Name, art.SortName, art.MbzArtistID, art.CreatedAt, art.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &art, nil
}

func (r *CatalogRepository) GetArtistByID(ctx context.Context, id string) (*domain.Artist, error) {
	query := `SELECT id, name, sort_name, mbz_artist_id, biography, album_count, track_count, created_at, updated_at
		FROM artists WHERE id = ?`
	var art domain.Artist
	var mbz, bio sql.NullString
	err := r.db.Executor().QueryRowContext(ctx, query, id).Scan(
		&art.ID, &art.Name, &art.SortName, &mbz, &bio, &art.AlbumCount, &art.TrackCount, &art.CreatedAt, &art.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("Artist", id)
		}
		return nil, err
	}
	if mbz.Valid {
		art.MbzArtistID = mbz.String
	}
	if bio.Valid {
		art.Biography = bio.String
	}
	return &art, nil
}

func (r *CatalogRepository) ListArtists(ctx context.Context, limit, offset int) ([]*domain.Artist, int, error) {
	var total int
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM artists`).Scan(&total)

	query := `SELECT id, name, sort_name, mbz_artist_id, album_count, track_count, created_at, updated_at
		FROM artists ORDER BY sort_name COLLATE NOCASE ASC LIMIT ? OFFSET ?`
	rows, err := r.db.Executor().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var artists []*domain.Artist
	for rows.Next() {
		var a domain.Artist
		var mbz sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &a.SortName, &mbz, &a.AlbumCount, &a.TrackCount, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if mbz.Valid {
			a.MbzArtistID = mbz.String
		}
		artists = append(artists, &a)
	}
	return artists, total, nil
}

// ----------------- Album Operations -----------------

func (r *CatalogRepository) FindOrCreateAlbum(ctx context.Context, exec DBExecutor, album *domain.Album) (*domain.Album, error) {
	var existing domain.Album
	query := `SELECT id, title, sort_title, album_artist_id, year, original_year, release_date, disc_count, track_count, duration, size_bytes, is_compilation, mbz_album_id, created_at, updated_at
		FROM albums WHERE title = ? COLLATE NOCASE AND album_artist_id = ?`
	var relDate, mbz sql.NullString
	err := exec.QueryRowContext(ctx, query, album.Title, album.AlbumArtistID).Scan(
		&existing.ID, &existing.Title, &existing.SortTitle, &existing.AlbumArtistID, &existing.Year, &existing.OriginalYear, &relDate,
		&existing.DiscCount, &existing.TrackCount, &existing.Duration, &existing.SizeBytes, &existing.IsCompilation, &mbz,
		&existing.CreatedAt, &existing.UpdatedAt,
	)
	if err == nil {
		if relDate.Valid {
			existing.ReleaseDate = relDate.String
		}
		if mbz.Valid {
			existing.MbzAlbumID = mbz.String
		}
		return &existing, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	album.ID = uuid.NewString()
	album.CreatedAt = time.Now().UTC()
	album.UpdatedAt = time.Now().UTC()

	insertQuery := `INSERT INTO albums (id, title, sort_title, album_artist_id, year, original_year, release_date, disc_count, track_count, duration, size_bytes, is_compilation, mbz_album_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = exec.ExecContext(ctx, insertQuery,
		album.ID, album.Title, album.SortTitle, album.AlbumArtistID, album.Year, album.OriginalYear, album.ReleaseDate,
		album.DiscCount, album.TrackCount, album.Duration, album.SizeBytes, album.IsCompilation, album.MbzAlbumID, album.CreatedAt, album.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return album, nil
}

func (r *CatalogRepository) GetAlbumByID(ctx context.Context, id string) (*domain.Album, error) {
	query := `SELECT a.id, a.title, a.sort_title, a.album_artist_id, ar.name, a.year, a.original_year, a.release_date, a.disc_count, a.track_count, a.duration, a.size_bytes, a.is_compilation, a.mbz_album_id, a.created_at, a.updated_at
		FROM albums a
		JOIN artists ar ON a.album_artist_id = ar.id
		WHERE a.id = ?`
	var alb domain.Album
	var relDate, mbz sql.NullString
	err := r.db.Executor().QueryRowContext(ctx, query, id).Scan(
		&alb.ID, &alb.Title, &alb.SortTitle, &alb.AlbumArtistID, &alb.AlbumArtist, &alb.Year, &alb.OriginalYear, &relDate,
		&alb.DiscCount, &alb.TrackCount, &alb.Duration, &alb.SizeBytes, &alb.IsCompilation, &mbz, &alb.CreatedAt, &alb.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("Album", id)
		}
		return nil, err
	}
	if relDate.Valid {
		alb.ReleaseDate = relDate.String
	}
	if mbz.Valid {
		alb.MbzAlbumID = mbz.String
	}
	return &alb, nil
}

func (r *CatalogRepository) ListAlbums(ctx context.Context, limit, offset int) ([]*domain.Album, int, error) {
	var total int
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM albums`).Scan(&total)

	query := `SELECT a.id, a.title, a.sort_title, a.album_artist_id, ar.name, a.year, a.original_year, a.release_date, a.disc_count, a.track_count, a.duration, a.size_bytes, a.is_compilation, a.mbz_album_id, a.created_at, a.updated_at
		FROM albums a
		JOIN artists ar ON a.album_artist_id = ar.id
		ORDER BY a.sort_title COLLATE NOCASE ASC LIMIT ? OFFSET ?`
	rows, err := r.db.Executor().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var albums []*domain.Album
	for rows.Next() {
		var a domain.Album
		var relDate, mbz sql.NullString
		if err := rows.Scan(
			&a.ID, &a.Title, &a.SortTitle, &a.AlbumArtistID, &a.AlbumArtist, &a.Year, &a.OriginalYear, &relDate,
			&a.DiscCount, &a.TrackCount, &a.Duration, &a.SizeBytes, &a.IsCompilation, &mbz, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if relDate.Valid {
			a.ReleaseDate = relDate.String
		}
		if mbz.Valid {
			a.MbzAlbumID = mbz.String
		}
		albums = append(albums, &a)
	}
	return albums, total, nil
}

// ----------------- Track Operations -----------------

func (r *CatalogRepository) UpsertTrack(ctx context.Context, exec DBExecutor, t *domain.Track) error {
	now := time.Now().UTC()
	t.UpdatedAt = now
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.ID == "" {
		t.ID = uuid.NewString()
	}

	query := `INSERT INTO tracks (
		id, pid, library_id, path, folder_path, filename, title, sort_title, raw_artist, album_id,
		track_number, disc_number, disc_subtitle, year, duration, bit_rate, sample_rate, bit_depth,
		channels, format, codec, file_size, rg_track_gain, rg_track_peak, rg_album_gain, rg_album_peak,
		has_embedded_cover, mbz_track_id, mtime, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		path = excluded.path,
		folder_path = excluded.folder_path,
		filename = excluded.filename,
		title = excluded.title,
		sort_title = excluded.sort_title,
		raw_artist = excluded.raw_artist,
		album_id = excluded.album_id,
		track_number = excluded.track_number,
		disc_number = excluded.disc_number,
		disc_subtitle = excluded.disc_subtitle,
		year = excluded.year,
		duration = excluded.duration,
		bit_rate = excluded.bit_rate,
		sample_rate = excluded.sample_rate,
		bit_depth = excluded.bit_depth,
		channels = excluded.channels,
		format = excluded.format,
		codec = excluded.codec,
		file_size = excluded.file_size,
		rg_track_gain = excluded.rg_track_gain,
		rg_track_peak = excluded.rg_track_peak,
		rg_album_gain = excluded.rg_album_gain,
		rg_album_peak = excluded.rg_album_peak,
		has_embedded_cover = excluded.has_embedded_cover,
		mbz_track_id = excluded.mbz_track_id,
		mtime = excluded.mtime,
		updated_at = excluded.updated_at`

	_, err := exec.ExecContext(ctx, query,
		t.ID, t.PID, t.LibraryID, t.Path, t.FolderPath, t.Filename, t.Title, t.SortTitle, t.RawArtist, t.AlbumID,
		t.TrackNumber, t.DiscNumber, t.DiscSubtitle, t.Year, t.Duration, t.BitRate, t.SampleRate, t.BitDepth,
		t.Channels, t.Format, t.Codec, t.FileSize, t.RGTrackGain, t.RGTrackPeak, t.RGAlbumGain, t.RGAlbumPeak,
		t.HasEmbeddedCover, t.MbzTrackID, t.MTime, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) GetTrackByID(ctx context.Context, id string) (*domain.Track, error) {
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		WHERE t.id = ?`

	var t domain.Track
	var discSub, mbz sql.NullString
	err := r.db.Executor().QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle, &t.RawArtist, &t.AlbumID,
		&t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub, &t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth,
		&t.Channels, &t.Format, &t.Codec, &t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
		&t.HasEmbeddedCover, &mbz, &t.MTime, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("Track", id)
		}
		return nil, err
	}
	if discSub.Valid {
		t.DiscSubtitle = discSub.String
	}
	if mbz.Valid {
		t.MbzTrackID = mbz.String
	}
	return &t, nil
}

func (r *CatalogRepository) ListTracksByAlbum(ctx context.Context, albumID string) ([]*domain.Track, error) {
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		WHERE t.album_id = ?
		ORDER BY t.disc_number ASC, t.track_number ASC`

	rows, err := r.db.Executor().QueryContext(ctx, query, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*domain.Track
	for rows.Next() {
		var t domain.Track
		var discSub, mbz sql.NullString
		if err := rows.Scan(
			&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle, &t.RawArtist, &t.AlbumID,
			&t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub, &t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth,
			&t.Channels, &t.Format, &t.Codec, &t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
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
		tracks = append(tracks, &t)
	}
	return tracks, nil
}

func (r *CatalogRepository) ListTracks(ctx context.Context, limit, offset int) ([]*domain.Track, int, error) {
	var total int
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&total)

	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		ORDER BY t.sort_title COLLATE NOCASE ASC LIMIT ? OFFSET ?`

	rows, err := r.db.Executor().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tracks []*domain.Track
	for rows.Next() {
		var t domain.Track
		var discSub, mbz sql.NullString
		if err := rows.Scan(
			&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle, &t.RawArtist, &t.AlbumID,
			&t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub, &t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth,
			&t.Channels, &t.Format, &t.Codec, &t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
			&t.HasEmbeddedCover, &mbz, &t.MTime, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if discSub.Valid {
			t.DiscSubtitle = discSub.String
		}
		if mbz.Valid {
			t.MbzTrackID = mbz.String
		}
		tracks = append(tracks, &t)
	}
	return tracks, total, nil
}

func (r *CatalogRepository) FindTrackByPath(ctx context.Context, path string) (trackID string, pid string, mtime int64, err error) {
	query := `SELECT id, pid, mtime FROM tracks WHERE path = ?`
	err = r.db.Executor().QueryRowContext(ctx, query, path).Scan(&trackID, &pid, &mtime)
	return
}

func (r *CatalogRepository) FindTrackByMBID(ctx context.Context, mbid string) (trackID string, pid string, err error) {
	query := `SELECT id, pid FROM tracks WHERE mbz_track_id = ? LIMIT 1`
	err = r.db.Executor().QueryRowContext(ctx, query, mbid).Scan(&trackID, &pid)
	return
}

func (r *CatalogRepository) DeleteTrackByPath(ctx context.Context, exec DBExecutor, path string) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM tracks WHERE path = ?`, path)
	return err
}

// ----------------- Track Artists Join Operations -----------------

func (r *CatalogRepository) SyncTrackArtists(ctx context.Context, exec DBExecutor, trackID string, artists []domain.ParsedArtist) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM track_artists WHERE track_id = ?`, trackID)
	if err != nil {
		return err
	}

	for idx, pa := range artists {
		name := strings.TrimSpace(pa.Name)
		if name == "" {
			continue
		}
		art, err := r.FindOrCreateArtist(ctx, exec, name, name, "")
		if err != nil {
			return err
		}
		_, err = exec.ExecContext(ctx, `INSERT OR REPLACE INTO track_artists (track_id, artist_id, role, position) VALUES (?, ?, ?, ?)`,
			trackID, art.ID, string(pa.Role), idx,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// ----------------- Genres Operations -----------------

func (r *CatalogRepository) FindOrCreateGenre(ctx context.Context, exec DBExecutor, name string) (*domain.Genre, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Unknown"
	}
	var g domain.Genre
	err := exec.QueryRowContext(ctx, `SELECT id, name, track_count, album_count FROM genres WHERE name = ? COLLATE NOCASE`, name).Scan(
		&g.ID, &g.Name, &g.TrackCount, &g.AlbumCount,
	)
	if err == nil {
		return &g, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	g = domain.Genre{
		ID:   uuid.NewString(),
		Name: name,
	}
	_, err = exec.ExecContext(ctx, `INSERT INTO genres (id, name) VALUES (?, ?)`, g.ID, g.Name)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *CatalogRepository) SyncTrackGenres(ctx context.Context, exec DBExecutor, trackID string, genreNames []string) error {
	_, _ = exec.ExecContext(ctx, `DELETE FROM track_genres WHERE track_id = ?`, trackID)
	for _, gn := range genreNames {
		gn = strings.TrimSpace(gn)
		if gn == "" {
			continue
		}
		g, err := r.FindOrCreateGenre(ctx, exec, gn)
		if err != nil {
			continue
		}
		_, _ = exec.ExecContext(ctx, `INSERT OR IGNORE INTO track_genres (track_id, genre_id) VALUES (?, ?)`, trackID, g.ID)
	}
	return nil
}

func (r *CatalogRepository) ListGenres(ctx context.Context) ([]*domain.Genre, error) {
	query := `SELECT id, name, track_count, album_count FROM genres ORDER BY name COLLATE NOCASE ASC`
	rows, err := r.db.Executor().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*domain.Genre
	for rows.Next() {
		var g domain.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.TrackCount, &g.AlbumCount); err != nil {
			return nil, err
		}
		genres = append(genres, &g)
	}
	return genres, nil
}

// ----------------- FTS5 Search & Statistics -----------------

func (r *CatalogRepository) IndexTrackFTS(ctx context.Context, exec DBExecutor, trackID, title, artist, album, genre string) error {
	_, _ = exec.ExecContext(ctx, `DELETE FROM track_search_fts WHERE track_id = ?`, trackID)
	query := `INSERT INTO track_search_fts (track_id, title, artist, album, genre) VALUES (?, ?, ?, ?, ?)`
	_, err := exec.ExecContext(ctx, query, trackID, title, artist, album, genre)
	return err
}

func (r *CatalogRepository) RecalculateAllStats(ctx context.Context) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// Update Album counts and durations
		_, err := tx.ExecContext(ctx, `UPDATE albums SET
			track_count = (SELECT COUNT(*) FROM tracks WHERE tracks.album_id = albums.id),
			duration = (SELECT COALESCE(SUM(duration), 0) FROM tracks WHERE tracks.album_id = albums.id),
			size_bytes = (SELECT COALESCE(SUM(file_size), 0) FROM tracks WHERE tracks.album_id = albums.id)`)
		if err != nil {
			return err
		}

		// Update Artist counts
		_, err = tx.ExecContext(ctx, `UPDATE artists SET
			album_count = (SELECT COUNT(*) FROM albums WHERE albums.album_artist_id = artists.id),
			track_count = (SELECT COUNT(DISTINCT track_id) FROM track_artists WHERE track_artists.artist_id = artists.id)`)
		if err != nil {
			return err
		}

		// Update Genre counts
		_, err = tx.ExecContext(ctx, `UPDATE genres SET
			track_count = (SELECT COUNT(*) FROM track_genres WHERE track_genres.genre_id = genres.id),
			album_count = (SELECT COUNT(*) FROM album_genres WHERE album_genres.genre_id = genres.id)`)
		return err
	})
}

// Search executes full-text search against track_search_fts and performs prefix matching against artists and albums.
func (r *CatalogRepository) Search(ctx context.Context, queryStr string, limit int) (*domain.SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	cleanQuery := strings.TrimSpace(queryStr)
	if cleanQuery == "" {
		return &domain.SearchResult{
			Query:   queryStr,
			Artists: []*domain.Artist{},
			Albums:  []*domain.Album{},
			Tracks:  []*domain.Track{},
		}, nil
	}

	result := &domain.SearchResult{
		Query:   cleanQuery,
		Artists: []*domain.Artist{},
		Albums:  []*domain.Album{},
		Tracks:  []*domain.Track{},
	}

	// 1. Search Artists
	likeQuery := "%" + cleanQuery + "%"
	artRows, err := r.db.Executor().QueryContext(ctx,
		`SELECT id, name, sort_name, mbz_artist_id, album_count, track_count, created_at, updated_at
		FROM artists WHERE name LIKE ? ORDER BY track_count DESC LIMIT ?`, likeQuery, limit)
	if err == nil {
		defer artRows.Close()
		for artRows.Next() {
			var a domain.Artist
			var mbz sql.NullString
			if err := artRows.Scan(&a.ID, &a.Name, &a.SortName, &mbz, &a.AlbumCount, &a.TrackCount, &a.CreatedAt, &a.UpdatedAt); err == nil {
				if mbz.Valid {
					a.MbzArtistID = mbz.String
				}
				result.Artists = append(result.Artists, &a)
			}
		}
	}

	// 2. Search Albums
	albRows, err := r.db.Executor().QueryContext(ctx,
		`SELECT a.id, a.title, a.sort_title, a.album_artist_id, ar.name, a.year, a.original_year,
		a.disc_count, a.track_count, a.duration, a.size_bytes, a.is_compilation, a.mbz_album_id, a.created_at, a.updated_at
		FROM albums a
		LEFT JOIN artists ar ON a.album_artist_id = ar.id
		WHERE a.title LIKE ? ORDER BY a.year DESC LIMIT ?`, likeQuery, limit)
	if err == nil {
		defer albRows.Close()
		for albRows.Next() {
			var a domain.Album
			var mbz sql.NullString
			if err := albRows.Scan(&a.ID, &a.Title, &a.SortTitle, &a.AlbumArtistID, &a.AlbumArtist, &a.Year, &a.OriginalYear,
				&a.DiscCount, &a.TrackCount, &a.Duration, &a.SizeBytes, &a.IsCompilation, &mbz, &a.CreatedAt, &a.UpdatedAt); err == nil {
				if mbz.Valid {
					a.MbzAlbumID = mbz.String
				}
				result.Albums = append(result.Albums, &a)
			}
		}
	}

	// 3. Search Tracks via FTS5
	// Sanitize query for FTS5: strip punctuation and wrap each token in quotes with prefix matching
	sanitized := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`"':*^()[]{}-+`, r) {
			return ' '
		}
		return r
	}, cleanQuery)

	words := strings.Fields(sanitized)
	if len(words) == 0 {
		return result, nil
	}

	var ftsParts []string
	for _, w := range words {
		if w != "" {
			ftsParts = append(ftsParts, w+"*")
		}
	}
	ftsQuery := strings.Join(ftsParts, " ")

	ftsRows, err := r.db.Executor().QueryContext(ctx,
		`SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title,
		t.raw_artist, t.album_id, al.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle,
		t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth, t.channels, t.format, t.codec,
		t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM track_search_fts fts
		JOIN tracks t ON fts.track_id = t.id
		JOIN albums al ON t.album_id = al.id
		LEFT JOIN artists ar ON al.album_artist_id = ar.id
		WHERE track_search_fts MATCH ?
		LIMIT ?`, ftsQuery, limit)
	if err == nil {
		defer ftsRows.Close()
		for ftsRows.Next() {
			var t domain.Track
			var discSub, mbz sql.NullString
			if err := ftsRows.Scan(
				&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle,
				&t.RawArtist, &t.AlbumID, &t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub,
				&t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth, &t.Channels, &t.Format, &t.Codec,
				&t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
				&t.HasEmbeddedCover, &mbz, &t.MTime, &t.CreatedAt, &t.UpdatedAt,
			); err == nil {
				if discSub.Valid {
					t.DiscSubtitle = discSub.String
				}
				if mbz.Valid {
					t.MbzTrackID = mbz.String
				}
				result.Tracks = append(result.Tracks, &t)
			}
		}
	}

	// Fallback to LIKE if FTS returned 0
	if len(result.Tracks) == 0 {
		fallbackRows, err := r.db.Executor().QueryContext(ctx,
			`SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title,
			t.raw_artist, t.album_id, al.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle,
			t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth, t.channels, t.format, t.codec,
			t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
			t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
			FROM tracks t
			JOIN albums al ON t.album_id = al.id
			LEFT JOIN artists ar ON al.album_artist_id = ar.id
			WHERE t.title LIKE ? OR t.raw_artist LIKE ?
			LIMIT ?`, likeQuery, likeQuery, limit)
		if err == nil {
			defer fallbackRows.Close()
			for fallbackRows.Next() {
				var t domain.Track
				var discSub, mbz sql.NullString
				if err := fallbackRows.Scan(
					&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle,
					&t.RawArtist, &t.AlbumID, &t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub,
					&t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth, &t.Channels, &t.Format, &t.Codec,
					&t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
					&t.HasEmbeddedCover, &mbz, &t.MTime, &t.CreatedAt, &t.UpdatedAt,
				); err == nil {
					if discSub.Valid {
						t.DiscSubtitle = discSub.String
					}
					if mbz.Valid {
						t.MbzTrackID = mbz.String
					}
					result.Tracks = append(result.Tracks, &t)
				}
			}
		}
	}

	return result, nil
}

type SystemStats struct {
	TotalTracks    int     `json:"totalTracks"`
	TotalAlbums    int     `json:"totalAlbums"`
	TotalArtists   int     `json:"totalArtists"`
	TotalGenres    int     `json:"totalGenres"`
	TotalPlaylists int     `json:"totalPlaylists"`
	TotalDuration  float64 `json:"totalDuration"`
	TotalBytes     int64   `json:"totalBytes"`
}

func (r *CatalogRepository) GetStats(ctx context.Context) (*SystemStats, error) {
	var stats SystemStats
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(duration), 0), COALESCE(SUM(file_size), 0) FROM tracks`).Scan(&stats.TotalTracks, &stats.TotalDuration, &stats.TotalBytes)
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM albums`).Scan(&stats.TotalAlbums)
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM artists`).Scan(&stats.TotalArtists)
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM genres`).Scan(&stats.TotalGenres)
	_ = r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM playlists`).Scan(&stats.TotalPlaylists)
	return &stats, nil
}

// ----------------- Smart Playlists -----------------

func (r *CatalogRepository) GetRecentlyAddedTracks(ctx context.Context, limit int) ([]*domain.Track, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		ORDER BY t.created_at DESC LIMIT ?`
	return r.scanTracksQuery(ctx, query, limit)
}

func (r *CatalogRepository) GetMostPlayedTracks(ctx context.Context, userID string, limit int) ([]*domain.Track, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		JOIN user_track_annotations an ON t.id = an.track_id
		WHERE an.user_id = ? AND an.play_count > 0
		ORDER BY an.play_count DESC, an.last_played_at DESC LIMIT ?`
	return r.scanTracksQuery(ctx, query, userID, limit)
}

func (r *CatalogRepository) GetRecentlyPlayedTracks(ctx context.Context, userID string, limit int) ([]*domain.Track, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		JOIN playback_history h ON t.id = h.track_id
		WHERE h.user_id = ?
		GROUP BY t.id
		ORDER BY MAX(h.played_at) DESC LIMIT ?`
	return r.scanTracksQuery(ctx, query, userID, limit)
}

func (r *CatalogRepository) GetTopRatedTracks(ctx context.Context, userID string, limit int) ([]*domain.Track, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		JOIN user_track_annotations an ON t.id = an.track_id
		WHERE an.user_id = ? AND an.rating >= 4
		ORDER BY an.rating DESC, an.rated_at DESC LIMIT ?`
	return r.scanTracksQuery(ctx, query, userID, limit)
}

func (r *CatalogRepository) GetRandomTracks(ctx context.Context, limit int) ([]*domain.Track, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT t.id, t.pid, t.library_id, t.path, t.folder_path, t.filename, t.title, t.sort_title, t.raw_artist, t.album_id,
		a.title, ar.name, t.track_number, t.disc_number, t.disc_subtitle, t.year, t.duration, t.bit_rate, t.sample_rate, t.bit_depth,
		t.channels, t.format, t.codec, t.file_size, t.rg_track_gain, t.rg_track_peak, t.rg_album_gain, t.rg_album_peak,
		t.has_embedded_cover, t.mbz_track_id, t.mtime, t.created_at, t.updated_at
		FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		ORDER BY RANDOM() LIMIT ?`
	return r.scanTracksQuery(ctx, query, limit)
}

func (r *CatalogRepository) scanTracksQuery(ctx context.Context, query string, args ...any) ([]*domain.Track, error) {
	rows, err := r.db.Executor().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*domain.Track
	for rows.Next() {
		var t domain.Track
		var discSub, mbz sql.NullString
		if err := rows.Scan(
			&t.ID, &t.PID, &t.LibraryID, &t.Path, &t.FolderPath, &t.Filename, &t.Title, &t.SortTitle, &t.RawArtist, &t.AlbumID,
			&t.AlbumTitle, &t.AlbumArtist, &t.TrackNumber, &t.DiscNumber, &discSub, &t.Year, &t.Duration, &t.BitRate, &t.SampleRate, &t.BitDepth,
			&t.Channels, &t.Format, &t.Codec, &t.FileSize, &t.RGTrackGain, &t.RGTrackPeak, &t.RGAlbumGain, &t.RGAlbumPeak,
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
		tracks = append(tracks, &t)
	}
	return tracks, nil
}