package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ne/internal/auth"
	"ne/internal/config"
	"ne/internal/media"
	"ne/internal/repository"
	"ne/internal/scanner"
	"ne/internal/service"
	transport "ne/internal/transport/http"
	"ne/internal/transport/subsonic"
)

var (
	version = "1.0.0"
	commit  = "clean"
	date    = "2026-08-30"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("nE Audio Server v%s (commit: %s, built: %s)\n", version, commit, date)
			return
		case "check-db":
			runCheckDB()
			return
		case "rebuild-fts":
			runRebuildFTS()
			return
		case "recalculate-stats":
			runRecalculateStats()
			return
		case "backup":
			dest := "ne_backup.db"
			if len(os.Args) > 2 {
				dest = os.Args[2]
			}
			runBackup(dest)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		}
	}

	runServer()
}

func printHelp() {
	fmt.Printf(`nE — Autonomous Personal Music Server & Streaming Platform

Usage:
  ne [command] [flags]

Available Commands:
  (no args)           Start the nE audio server (HTTP API & Web Player)
  check-db            Perform database integrity check and verify foreign key constraints
  rebuild-fts         Rebuild FTS5 search index from all library tracks
  recalculate-stats   Recalculate aggregate stats for artists, albums, genres
  backup [dest]       Create a hot consistent copy of the SQLite database using Online Backup (VACUUM INTO)
  version             Show server version and build information
  help                Show this help message

Environment Variables:
  NE_PORT             HTTP port to listen on (default: 4533)
  NE_HOST             HTTP host address (default: 0.0.0.0)
  NE_DATA_DIR         Directory for SQLite database (default: ./data)
  NE_CONFIG_DIR       Directory for config files and keys (default: ./config)
  NE_CACHE_DIR        Directory for transcoded audio & artwork (default: ./cache)
  NE_MUSIC_DIR        Default music folder path
  NE_JWT_SECRET       Secret key for JWT token signing
  NE_LOG_LEVEL        Log level: debug, info, warn, error (default: info)
  NE_LOG_FORMAT       Log output format: text, json (default: text)
`)
}

func runServer() {
	// Load configuration
	configFile := ""
	if len(os.Args) > 2 && (os.Args[1] == "--config" || os.Args[1] == "-c") {
		configFile = os.Args[2]
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	var logHandler slog.Handler
	var logLevel slog.Level
	switch cfg.Server.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	if cfg.Server.LogFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(logHandler)

	logger.Info("Starting nE Music Server",
		"version", version,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"db", cfg.Database.Path,
	)

	// Open SQLite database
	db, err := repository.Open(cfg.Database.Path, cfg.Database.BusyTimeout, logger)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize JWT Manager
	jwtMgr, err := auth.NewJWTManager(
		cfg.Auth.JWTSecretKey,
		cfg.Paths.ConfigDir,
		cfg.Auth.AccessTokenExpiry,
		cfg.Auth.RefreshTokenExpiry,
	)
	if err != nil {
		logger.Error("Failed to initialize JWT manager", "error", err)
		os.Exit(1)
	}

	// Repositories & Services
	userRepo := repository.NewUserRepository(db)
	catalogRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)

	authService := service.NewAuthService(userRepo, jwtMgr)
	scannerEngine := scanner.NewScanner(db, catalogRepo, &cfg.Scanner, logger)
	catalogService := service.NewCatalogService(catalogRepo, scannerEngine)
	transcoderPool := media.NewTranscoder(cfg.Streaming.MaxConcurrentTranscodes)
	streamService := service.NewStreamService(catalogRepo, transcoderPool)
	artworkService := service.NewArtworkService(catalogRepo, cfg.Paths.CacheDir)
	annoService := service.NewAnnotationService(annoRepo)
	playlistService := service.NewPlaylistService(playlistRepo)
	smartPlaylistService := service.NewSmartPlaylistService(catalogRepo)
	lyricsService := service.NewLyricsService(catalogRepo, cfg.Paths.CacheDir)
	artistMetaService := service.NewArtistMetaService(catalogRepo)
	githubSyncService := service.NewGitHubSyncService(db, &cfg.GitHub, logger)
	subsonicHandler := subsonic.NewSubsonicHandler(userRepo, catalogService, streamService, artworkService, annoService, playlistService)

	// Auto-discovery: If no libraries exist, automatically detect and register default music folder
	go func() {
		time.Sleep(500 * time.Millisecond)
		bgCtx := context.Background()
		libs, err := catalogService.ListLibraries(bgCtx)
		if err == nil && len(libs) == 0 {
			musicDir := cfg.Paths.MusicDir
			if _, err := os.Stat(musicDir); err == nil {
				logger.Info("Auto-discovering music directory on startup", "path", musicDir)
				lib, err := catalogService.CreateLibrary(bgCtx, "Default Library", musicDir)
				if err == nil {
					logger.Info("Triggering startup scan for default library", "libraryId", lib.ID)
					_, _ = catalogService.TriggerScan(bgCtx, lib.ID)
				}
			}
		} else if len(libs) > 0 {
			stats, _ := catalogRepo.GetStats(bgCtx)
			if stats == nil || stats.TotalTracks == 0 || cfg.Scanner.ScanOnStartup {
				logger.Info("Triggering startup library scan", "libraryCount", len(libs))
				for _, l := range libs {
					_, _ = catalogService.TriggerScan(bgCtx, l.ID)
				}
			}
		}

		// Continuous Library Watcher (periodically inspects libraries for new/changed files)
		watcher := scanner.NewLibraryWatcher(catalogService, 20*time.Second, logger)
		watcher.Start(bgCtx)
	}()

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create and start HTTP server
	server := transport.NewServer(
		cfg,
		logger,
		func() error { return db.Ping(context.Background()) },
		jwtMgr,
		authService,
		catalogService,
		streamService,
		artworkService,
		annoService,
		playlistService,
		smartPlaylistService,
		lyricsService,
		artistMetaService,
		subsonicHandler,
		githubSyncService,
	)

	if err := server.Start(ctx); err != nil {
		logger.Error("Server stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("nE server shut down cleanly.")
}

func runCheckDB() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := repository.Open(cfg.Database.Path, 5000, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	var integrity string
	if err := db.Executor().QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		fmt.Fprintf(os.Stderr, "Integrity check failed to execute: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Database Integrity Check: %s\n", integrity)
	if integrity != "ok" {
		fmt.Fprintf(os.Stderr, "Database integrity violation detected!\n")
		os.Exit(1)
	}

	rows, err := db.Executor().QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Foreign key check failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	violations := 0
	for rows.Next() {
		violations++
		var table, parent string
		var rowid int64
		var fkid int
		if err := rows.Scan(&table, &rowid, &parent, &fkid); err == nil {
			fmt.Printf("Foreign key violation in table %s (rowid: %d, references: %s)\n", table, rowid, parent)
		}
	}
	if violations > 0 {
		fmt.Fprintf(os.Stderr, "Total Foreign Key Violations: %d\n", violations)
		os.Exit(1)
	}
	fmt.Println("Foreign Key Constraints: OK (0 violations)")
}

func runRebuildFTS() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := repository.Open(cfg.Database.Path, 5000, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	if _, err := db.Executor().ExecContext(ctx, `DELETE FROM track_search_fts`); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clear FTS index: %v\n", err)
		os.Exit(1)
	}
	_, err = db.Executor().ExecContext(ctx, `INSERT INTO track_search_fts (track_id, title, artist, album, genre)
		SELECT t.id, t.title, t.raw_artist, al.title, ''
		FROM tracks t
		JOIN albums al ON t.album_id = al.id`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rebuild FTS failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("FTS5 full-text search index rebuilt successfully!")
}

func runRecalculateStats() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := repository.Open(cfg.Database.Path, 5000, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewCatalogRepository(db)
	if err := repo.RecalculateAllStats(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Recalculate stats failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Catalog statistics recalculated successfully!")
}

func runBackup(dest string) {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := repository.Open(cfg.Database.Path, 5000, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening source database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Backup(ctx, dest); err != nil {
		fmt.Fprintf(os.Stderr, "SQLite online backup failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Database online backup completed successfully to %q\n", dest)
}
