-- Schema Migration 001: Initial Schema for nE

PRAGMA foreign_keys = ON;

-- Schema Version Tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Libraries
CREATE TABLE IF NOT EXISTS libraries (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    last_scanned_at TIMESTAMP,
    scan_status TEXT NOT NULL DEFAULT 'idle',
    track_count INTEGER NOT NULL DEFAULT 0,
    album_count INTEGER NOT NULL DEFAULT 0,
    total_bytes INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE COLLATE NOCASE,
    email TEXT UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT 0,
    can_transcode BOOLEAN NOT NULL DEFAULT 1,
    token_version INTEGER NOT NULL DEFAULT 1,
    subsonic_salt TEXT,
    subsonic_token TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Hashed Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);

-- Artists
CREATE TABLE IF NOT EXISTS artists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sort_name TEXT NOT NULL,
    mbz_artist_id TEXT,
    biography TEXT,
    album_count INTEGER NOT NULL DEFAULT 0,
    track_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_artists_name ON artists(name COLLATE NOCASE);
CREATE INDEX IF NOT EXISTS idx_artists_sort_name ON artists(sort_name COLLATE NOCASE);

-- Albums
CREATE TABLE IF NOT EXISTS albums (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    sort_title TEXT NOT NULL,
    album_artist_id TEXT NOT NULL,
    year INTEGER NOT NULL DEFAULT 0,
    original_year INTEGER NOT NULL DEFAULT 0,
    release_date TEXT,
    disc_count INTEGER NOT NULL DEFAULT 1,
    track_count INTEGER NOT NULL DEFAULT 0,
    duration REAL NOT NULL DEFAULT 0,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    is_compilation BOOLEAN NOT NULL DEFAULT 0,
    mbz_album_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(album_artist_id) REFERENCES artists(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_albums_artist ON albums(album_artist_id);
CREATE INDEX IF NOT EXISTS idx_albums_title ON albums(title COLLATE NOCASE);
CREATE INDEX IF NOT EXISTS idx_albums_year ON albums(year);

-- Tracks
CREATE TABLE IF NOT EXISTS tracks (
    id TEXT PRIMARY KEY,
    pid TEXT NOT NULL UNIQUE,
    library_id TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    folder_path TEXT NOT NULL,
    filename TEXT NOT NULL,
    title TEXT NOT NULL,
    sort_title TEXT NOT NULL,
    raw_artist TEXT NOT NULL,
    album_id TEXT NOT NULL,
    track_number INTEGER NOT NULL DEFAULT 1,
    disc_number INTEGER NOT NULL DEFAULT 1,
    disc_subtitle TEXT,
    year INTEGER NOT NULL DEFAULT 0,
    duration REAL NOT NULL DEFAULT 0,
    bit_rate INTEGER NOT NULL DEFAULT 0,
    sample_rate INTEGER NOT NULL DEFAULT 0,
    bit_depth INTEGER,
    channels INTEGER NOT NULL DEFAULT 2,
    format TEXT NOT NULL,
    codec TEXT NOT NULL,
    file_size INTEGER NOT NULL DEFAULT 0,
    rg_track_gain REAL,
    rg_track_peak REAL,
    rg_album_gain REAL,
    rg_album_peak REAL,
    has_embedded_cover BOOLEAN NOT NULL DEFAULT 0,
    mbz_track_id TEXT,
    mtime INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(library_id) REFERENCES libraries(id) ON DELETE CASCADE,
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_tracks_library ON tracks(library_id);
CREATE INDEX IF NOT EXISTS idx_tracks_album ON tracks(album_id);
CREATE INDEX IF NOT EXISTS idx_tracks_pid ON tracks(pid);
CREATE INDEX IF NOT EXISTS idx_tracks_title ON tracks(title COLLATE NOCASE);

-- Track Artists
CREATE TABLE IF NOT EXISTS track_artists (
    track_id TEXT NOT NULL,
    artist_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'primary',
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(track_id, artist_id, role),
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY(artist_id) REFERENCES artists(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_track_artists_artist ON track_artists(artist_id);

-- Genres
CREATE TABLE IF NOT EXISTS genres (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    track_count INTEGER NOT NULL DEFAULT 0,
    album_count INTEGER NOT NULL DEFAULT 0
);

-- Track Genres
CREATE TABLE IF NOT EXISTS track_genres (
    track_id TEXT NOT NULL,
    genre_id TEXT NOT NULL,
    PRIMARY KEY(track_id, genre_id),
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY(genre_id) REFERENCES genres(id) ON DELETE CASCADE
);

-- Album Genres
CREATE TABLE IF NOT EXISTS album_genres (
    album_id TEXT NOT NULL,
    genre_id TEXT NOT NULL,
    PRIMARY KEY(album_id, genre_id),
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE,
    FOREIGN KEY(genre_id) REFERENCES genres(id) ON DELETE CASCADE
);

-- Explicit Referentially-Intact User Annotations
CREATE TABLE IF NOT EXISTS user_track_annotations (
    user_id TEXT NOT NULL,
    track_id TEXT NOT NULL,
    is_starred BOOLEAN NOT NULL DEFAULT 0,
    starred_at TIMESTAMP,
    rating INTEGER NOT NULL DEFAULT 0 CHECK(rating >= 0 AND rating <= 5),
    rated_at TIMESTAMP,
    play_count INTEGER NOT NULL DEFAULT 0,
    last_played_at TIMESTAMP,
    PRIMARY KEY(user_id, track_id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_track_starred ON user_track_annotations(user_id, is_starred);

CREATE TABLE IF NOT EXISTS user_album_annotations (
    user_id TEXT NOT NULL,
    album_id TEXT NOT NULL,
    is_starred BOOLEAN NOT NULL DEFAULT 0,
    starred_at TIMESTAMP,
    rating INTEGER NOT NULL DEFAULT 0 CHECK(rating >= 0 AND rating <= 5),
    rated_at TIMESTAMP,
    play_count INTEGER NOT NULL DEFAULT 0,
    last_played_at TIMESTAMP,
    PRIMARY KEY(user_id, album_id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_artist_annotations (
    user_id TEXT NOT NULL,
    artist_id TEXT NOT NULL,
    is_starred BOOLEAN NOT NULL DEFAULT 0,
    starred_at TIMESTAMP,
    rating INTEGER NOT NULL DEFAULT 0 CHECK(rating >= 0 AND rating <= 5),
    PRIMARY KEY(user_id, artist_id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(artist_id) REFERENCES artists(id) ON DELETE CASCADE
);

-- Playlists
CREATE TABLE IF NOT EXISTS playlists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    comment TEXT,
    owner_id TEXT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT 0,
    duration REAL NOT NULL DEFAULT 0,
    track_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(owner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Playlist Tracks
CREATE TABLE IF NOT EXISTS playlist_tracks (
    id TEXT PRIMARY KEY,
    playlist_id TEXT NOT NULL,
    track_id TEXT NOT NULL,
    position INTEGER NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    UNIQUE(playlist_id, position)
);
CREATE INDEX IF NOT EXISTS idx_playlist_tracks_order ON playlist_tracks(playlist_id, position);

-- Playback History
CREATE TABLE IF NOT EXISTS playback_history (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    track_id TEXT NOT NULL,
    player_name TEXT,
    played_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    duration_played REAL NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_history_user ON playback_history(user_id, played_at DESC);

-- Persistent Scan Runs
CREATE TABLE IF NOT EXISTS scan_runs (
    id TEXT PRIMARY KEY,
    library_id TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    files_found INTEGER NOT NULL DEFAULT 0,
    files_processed INTEGER NOT NULL DEFAULT 0,
    files_failed INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    FOREIGN KEY(library_id) REFERENCES libraries(id) ON DELETE CASCADE
);

-- FTS5 Search Virtual Table
CREATE VIRTUAL TABLE IF NOT EXISTS track_search_fts USING fts5(
    track_id UNINDEXED,
    title,
    artist,
    album,
    genre,
    tokenize='unicode61 remove_diacritics 2'
);
