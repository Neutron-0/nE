package scanner

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ne/internal/domain"
	"ne/internal/media"
)

type CatalogScanner interface {
	ListLibraries(ctx context.Context) ([]*domain.Library, error)
	TriggerScan(ctx context.Context, libraryID string) (*ScanProgress, error)
}

type LibraryWatcher struct {
	scanner   CatalogScanner
	interval  time.Duration
	logger    *slog.Logger
	lastState map[string]int // libraryID -> file count
	mu        sync.Mutex
	stopCh    chan struct{}
}

func NewLibraryWatcher(scanner CatalogScanner, interval time.Duration, logger *slog.Logger) *LibraryWatcher {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &LibraryWatcher{
		scanner:   scanner,
		interval:  interval,
		logger:    logger,
		lastState: make(map[string]int),
		stopCh:    make(chan struct{}),
	}
}

func (w *LibraryWatcher) Start(ctx context.Context) {
	w.logger.Info("Starting continuous music library watcher", "interval", w.interval)
	ticker := time.NewTicker(w.interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-ticker.C:
				w.checkLibraries(ctx)
			}
		}
	}()
}

func (w *LibraryWatcher) Stop() {
	close(w.stopCh)
}

func (w *LibraryWatcher) checkLibraries(ctx context.Context) {
	libs, err := w.scanner.ListLibraries(ctx)
	if err != nil || len(libs) == 0 {
		return
	}

	for _, lib := range libs {
		currentCount := w.countAudioFiles(lib.Path)
		w.mu.Lock()
		prevCount, exists := w.lastState[lib.ID]
		w.lastState[lib.ID] = currentCount
		w.mu.Unlock()

		if !exists {
			// First observation: if track count in library differs, trigger scan
			if currentCount > 0 && lib.TrackCount == 0 {
				w.logger.Info("LibraryWatcher: Detected unindexed tracks on boot, triggering scan", "library", lib.Name, "files", currentCount)
				_, _ = w.scanner.TriggerScan(ctx, lib.ID)
			}
		} else if currentCount != prevCount {
			w.logger.Info("LibraryWatcher: Detected file changes in library folder", "library", lib.Name, "oldCount", prevCount, "newCount", currentCount)
			_, _ = w.scanner.TriggerScan(ctx, lib.ID)
		}
	}
}

func (w *LibraryWatcher) countAudioFiles(root string) int {
	// Candidate fallbacks if path is relative or Linux/Windows mismatch
	resolved := root
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		candidates := []string{"music", "/music", filepath.Join(".", "music")}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				resolved = c
				break
			}
		}
	}

	count := 0
	_ = filepath.WalkDir(resolved, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if media.IsSupportedAudioFile(path) {
			count++
		}
		return nil
	})
	return count
}
