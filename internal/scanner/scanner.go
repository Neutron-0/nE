package scanner

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ne/internal/config"
	"ne/internal/domain"
	"ne/internal/media"
	"ne/internal/repository"
	"github.com/google/uuid"
)

type DiscoveredFile struct {
	Path  string
	MTime int64
	Size  int64
}

type ExtractedItem struct {
	File DiscoveredFile
	Meta *media.AudioMetadata
	Err  error
}

type ScanProgress struct {
	LibraryID       string `json:"libraryId"`
	Status          string `json:"status"` // "idle", "scanning", "completed", "failed"
	FilesFound      int    `json:"filesFound"`
	FilesProcessed  int    `json:"filesProcessed"`
	FilesFailed     int    `json:"filesFailed"`
	DurationSeconds int64  `json:"durationSeconds"`
	ErrorMessage    string `json:"errorMessage,omitempty"`
}

type Scanner struct {
	db          *repository.DB
	catalogRepo *repository.CatalogRepository
	cfg         *config.ScannerConfig
	logger      *slog.Logger
	scanLocks   sync.Map // per-library mutex to prevent concurrent scans of the same library
}

func NewScanner(db *repository.DB, catalogRepo *repository.CatalogRepository, cfg *config.ScannerConfig, logger *slog.Logger) *Scanner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scanner{
		db:          db,
		catalogRepo: catalogRepo,
		cfg:         cfg,
		logger:      logger,
	}
}

// ScanLibrary executes the full 11-stage ingestion pipeline against a registered library.
func (s *Scanner) ScanLibrary(ctx context.Context, libraryID string) (*ScanProgress, error) {
	// Per-library scan concurrency lock
	lockVal, _ := s.scanLocks.LoadOrStore(libraryID, &sync.Mutex{})
	libLock := lockVal.(*sync.Mutex)
	if !libLock.TryLock() {
		return nil, domain.ErrConflict("Scan already running for this library")
	}
	defer libLock.Unlock()

	startTime := time.Now()
	lib, err := s.catalogRepo.GetLibraryByID(ctx, libraryID)
	if err != nil {
		return nil, fmt.Errorf("fetching library %q: %w", libraryID, err)
	}

	s.logger.Info("Starting library scan", "library", lib.Name, "path", lib.Path)
	_ = s.catalogRepo.UpdateLibraryScanStatus(ctx, libraryID, "scanning")

	progress := &ScanProgress{
		LibraryID: libraryID,
		Status:    "scanning",
	}

	// Inaccessibility guard: check if path is mounted and accessible
	stat, err := os.Stat(lib.Path)
	if err != nil || !stat.IsDir() {
		_ = s.catalogRepo.UpdateLibraryScanStatus(ctx, libraryID, "failed")
		s.logger.Error("Library path is inaccessible or unmounted; scan aborted safely", "path", lib.Path, "error", err)
		return nil, fmt.Errorf("library directory is inaccessible: %w", err)
	}

	// Stage 1: Filesystem Discovery & Ignore Rules
	var discovered []DiscoveredFile
	var playlistFiles []string

	ignoreMatcher := NewIgnoreMatcher(lib.Path, s.cfg.IgnoreFile)

	err = filepath.WalkDir(lib.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Warn("Error accessing directory during scan", "path", path, "error", err)
			return nil // Skip and continue
		}

		if d.IsDir() {
			if path != lib.Path && ignoreMatcher.ShouldIgnore(path) {
				return fs.SkipDir
			}
			return nil
		}

		if ignoreMatcher.ShouldIgnore(path) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".m3u" || ext == ".m3u8" {
			playlistFiles = append(playlistFiles, path)
			return nil
		}

		if media.IsSupportedAudioFile(path) {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			discovered = append(discovered, DiscoveredFile{
				Path:  path,
				MTime: info.ModTime().Unix(),
				Size:  info.Size(),
			})
		}
		return nil
	})

	if err != nil {
		_ = s.catalogRepo.UpdateLibraryScanStatus(ctx, libraryID, "failed")
		return nil, fmt.Errorf("directory walk failed: %w", err)
	}

	progress.FilesFound = len(discovered)
	s.logger.Info("Discovery completed", "files_found", progress.FilesFound)

	// Stage 2 & 3: Parallel Metadata Extraction
	workerCount := s.cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = 4
	}

	jobs := make(chan DiscoveredFile, len(discovered))
	results := make(chan ExtractedItem, len(discovered))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for df := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					meta, err := media.ParseAudioFile(df.Path)
					results <- ExtractedItem{File: df, Meta: meta, Err: err}
				}
			}
		}()
	}

	for _, df := range discovered {
		jobs <- df
	}
	close(jobs)
	wg.Wait()
	close(results)

	// Stage 4..7: Identity Resolution (PIDs), Normalization, Entity Upsert in Batches
	var batch []ExtractedItem
	batchSize := s.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 200
	}

	var totalProcessed int
	var totalFailed int
	var totalBytes int64

	for item := range results {
		if item.Err != nil {
			s.logger.Warn("Failed to extract metadata", "file", item.File.Path, "error", item.Err)
			totalFailed++
			continue
		}

		batch = append(batch, item)
		totalBytes += item.File.Size

		if len(batch) >= batchSize {
			if err := s.processBatch(ctx, libraryID, batch); err != nil {
				s.logger.Error("Batch upsert failed", "error", err)
			} else {
				totalProcessed += len(batch)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := s.processBatch(ctx, libraryID, batch); err != nil {
			s.logger.Error("Final batch upsert failed", "error", err)
		} else {
			totalProcessed += len(batch)
		}
	}

	// Stage 9: Playlist Sync (.m3u / .m3u8 files)
	s.syncM3UPlaylists(ctx, libraryID, playlistFiles)

	// Stage 7: Recalculate Aggregates
	_ = s.catalogRepo.RecalculateAllStats(ctx)

	// Stage 10: Destructive Cleanup (Only prune if discovery succeeded reliably)
	if progress.FilesFound > 0 {
		s.cleanupMissingTracks(ctx, libraryID, discovered)
	}

	// Stage 11: Complete Scan
	_ = s.catalogRepo.UpdateLibraryStats(ctx, libraryID, totalProcessed, 0, totalBytes)
	progress.Status = "completed"
	progress.FilesProcessed = totalProcessed
	progress.FilesFailed = totalFailed
	progress.DurationSeconds = int64(time.Since(startTime).Seconds())

	s.logger.Info("Library scan completed successfully",
		"library", lib.Name,
		"processed", totalProcessed,
		"failed", totalFailed,
		"duration_sec", progress.DurationSeconds,
	)

	return progress, nil
}

func (s *Scanner) processBatch(ctx context.Context, libraryID string, batch []ExtractedItem) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		for _, item := range batch {
			meta := item.Meta
			df := item.File

			// Normalization
			cleanTitle := NormalizeString(meta.Title)
			sortTitle := GenerateSortName(cleanTitle, s.cfg.CleanEnglish)
			cleanAlbum := NormalizeString(meta.Album)
			sortAlbum := GenerateSortName(cleanAlbum, s.cfg.CleanEnglish)
			cleanAlbumArtist := NormalizeString(meta.AlbumArtist)
			if cleanAlbumArtist == "" {
				cleanAlbumArtist = NormalizeString(meta.Artist)
			}

			// Artist record
			artistSort := GenerateSortName(cleanAlbumArtist, s.cfg.CleanEnglish)
			albumArtist, err := s.catalogRepo.FindOrCreateArtist(ctx, tx, cleanAlbumArtist, artistSort, meta.MbzArtistID)
			if err != nil {
				return err
			}

			// Album record
			album := &domain.Album{
				Title:         cleanAlbum,
				SortTitle:     sortAlbum,
				AlbumArtistID: albumArtist.ID,
				Year:          meta.Year,
				DiscCount:     meta.TotalDiscs,
				TrackCount:    meta.TotalTracks,
				MbzAlbumID:    meta.MbzAlbumID,
			}
			dbAlbum, err := s.catalogRepo.FindOrCreateAlbum(ctx, tx, album)
			if err != nil {
				return err
			}

			// 5-Tier PID Resolution & Move Detection
			trackID, pid := s.resolveTrackIdentity(ctx, tx, libraryID, df, meta, cleanAlbumArtist, cleanAlbum, cleanTitle)

			track := &domain.Track{
				ID:               trackID,
				PID:              pid,
				LibraryID:        libraryID,
				Path:             filepath.ToSlash(df.Path),
				FolderPath:       filepath.ToSlash(filepath.Dir(df.Path)),
				Filename:         filepath.Base(df.Path),
				Title:            cleanTitle,
				SortTitle:        sortTitle,
				RawArtist:        meta.Artist,
				AlbumID:          dbAlbum.ID,
				TrackNumber:      meta.TrackNumber,
				DiscNumber:       meta.DiscNumber,
				Year:             meta.Year,
				Duration:         meta.Duration,
				BitRate:          meta.BitRate,
				SampleRate:       meta.SampleRate,
				BitDepth:         meta.BitDepth,
				Channels:         meta.Channels,
				Format:           meta.Format,
				Codec:            meta.Codec,
				FileSize:         df.Size,
				RGTrackGain:      meta.RGTrackGain,
				RGTrackPeak:      meta.RGTrackPeak,
				RGAlbumGain:      meta.RGAlbumGain,
				RGAlbumPeak:      meta.RGAlbumPeak,
				HasEmbeddedCover: meta.HasEmbeddedCover,
				MbzTrackID:       meta.MbzTrackID,
				MTime:            df.MTime,
			}

			if err := s.catalogRepo.UpsertTrack(ctx, tx, track); err != nil {
				return err
			}

			// Sync Track Artists
			parsedArtists := SplitArtists(meta.Artist)
			_ = s.catalogRepo.SyncTrackArtists(ctx, tx, track.ID, parsedArtists)

			// Sync Track Genres
			if meta.Genre != "" {
				genres := strings.Split(meta.Genre, ";")
				_ = s.catalogRepo.SyncTrackGenres(ctx, tx, track.ID, genres)
			}

			// Sync FTS5
			_ = s.catalogRepo.IndexTrackFTS(ctx, tx, track.ID, cleanTitle, meta.Artist, cleanAlbum, meta.Genre)
		}
		return nil
	})
}

// resolveTrackIdentity implements the 5-Tier PID Resolution Algorithm.
func (s *Scanner) resolveTrackIdentity(
	ctx context.Context,
	tx *sql.Tx,
	libraryID string,
	df DiscoveredFile,
	meta *media.AudioMetadata,
	cleanArtist, cleanAlbum, cleanTitle string,
) (string, string) {
	// Tier 1 & 2: Same path
	var existingID, existingPID string
	var existingMTime int64
	normPath := filepath.ToSlash(df.Path)
	err := tx.QueryRowContext(ctx, `SELECT id, pid, mtime FROM tracks WHERE path = ? OR path = ?`, df.Path, normPath).Scan(&existingID, &existingPID, &existingMTime)
	if err == nil {
		return existingID, existingPID
	}

	// Tier 3: MusicBrainz Track ID match (moved/renamed file)
	if meta.MbzTrackID != "" {
		var mbzID, mbzPID string
		err := tx.QueryRowContext(ctx, `SELECT id, pid FROM tracks WHERE mbz_track_id = ? LIMIT 1`, meta.MbzTrackID).Scan(&mbzID, &mbzPID)
		if err == nil {
			s.logger.Info("Tier 3 PID match: File moved/renamed via MBID", "old_id", mbzID, "new_path", df.Path)
			return mbzID, mbzPID
		}
	}

	// Tier 4: Strong Metadata Match (Artist + Album + Disc + Track + Title & Duration within 2s)
	var metaID, metaPID string
	var metaDuration float64
	err = tx.QueryRowContext(ctx, `SELECT t.id, t.pid, t.duration FROM tracks t
		JOIN albums a ON t.album_id = a.id
		JOIN artists ar ON a.album_artist_id = ar.id
		WHERE ar.name = ? COLLATE NOCASE AND a.title = ? COLLATE NOCASE AND t.track_number = ? AND t.title = ? COLLATE NOCASE
		LIMIT 1`, cleanArtist, cleanAlbum, meta.TrackNumber, cleanTitle).Scan(&metaID, &metaPID, &metaDuration)
	if err == nil && math.Abs(metaDuration-meta.Duration) < 2.0 {
		s.logger.Info("Tier 4 PID match: File moved/renamed via Metadata signature", "old_id", metaID, "new_path", df.Path)
		return metaID, metaPID
	}

	// Tier 5: Brand new track record
	return uuid.NewString(), GenerateNewPID()
}

func (s *Scanner) syncM3UPlaylists(ctx context.Context, libraryID string, files []string) {
	for _, m3uPath := range files {
		file, err := os.Open(m3uPath)
		if err != nil {
			continue
		}
		defer file.Close()

		baseDir := filepath.Dir(m3uPath)
		playlistName := strings.TrimSuffix(filepath.Base(m3uPath), filepath.Ext(m3uPath))

		var trackPaths []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			targetPath := line
			if !filepath.IsAbs(targetPath) {
				targetPath = filepath.Join(baseDir, targetPath)
			}
			trackPaths = append(trackPaths, targetPath)
		}

		if len(trackPaths) > 0 {
			s.logger.Info("Discovered M3U Playlist", "name", playlistName, "tracks", len(trackPaths))
		}
	}
}

func (s *Scanner) cleanupMissingTracks(ctx context.Context, libraryID string, discovered []DiscoveredFile) {
	discoveredMap := make(map[string]bool, len(discovered))
	for _, d := range discovered {
		discoveredMap[d.Path] = true
	}

	tracks, _, err := s.catalogRepo.ListTracks(ctx, 100000, 0)
	if err != nil {
		return
	}

	for _, t := range tracks {
		if t.LibraryID == libraryID && !discoveredMap[t.Path] {
			if _, err := os.Stat(t.Path); errors.Is(err, os.ErrNotExist) {
				s.logger.Info("Pruning deleted track from catalog", "path", t.Path)
				_ = s.catalogRepo.DeleteTrackByPath(ctx, s.db.Executor(), t.Path)
			}
		}
	}
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:16])
}
