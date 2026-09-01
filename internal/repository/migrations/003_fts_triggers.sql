-- Migration 003: Keep track_search_fts synchronized on track deletions and updates
CREATE TRIGGER IF NOT EXISTS trg_tracks_fts_delete AFTER DELETE ON tracks BEGIN
    DELETE FROM track_search_fts WHERE track_id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS trg_tracks_fts_update AFTER UPDATE OF title, raw_artist ON tracks BEGIN
    DELETE FROM track_search_fts WHERE track_id = OLD.id;
    INSERT INTO track_search_fts (track_id, title, artist, album, genre)
    SELECT NEW.id, NEW.title, NEW.raw_artist, a.title, ''
    FROM albums a WHERE a.id = NEW.album_id;
END;
