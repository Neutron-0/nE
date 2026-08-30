package repository

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DBExecutor abstracts queries across a direct *sql.DB connection or an active *sql.Tx.
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// DB wraps the SQLite connection and writer lock.
type DB struct {
	db      *sql.DB
	writeMu sync.Mutex
	logger  *slog.Logger
}

// Open initializes SQLite with WAL mode and runs pending migrations.
func Open(dbPath string, busyTimeout int, logger *slog.Logger) (*DB, error) {
	if logger == nil {
		logger = slog.Default()
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(%d)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath, busyTimeout)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database at %q: %w", dbPath, err)
	}

	// SQLite single-writer configuration
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("pinging sqlite database: %w", err)
	}

	d := &DB{
		db:     sqlDB,
		logger: logger,
	}

	if err := d.Migrate(context.Background()); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("running database migrations: %w", err)
	}

	logger.Info("SQLite database initialized successfully", "path", dbPath)
	return d, nil
}

// Close terminates the underlying database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// Ping checks database responsiveness.
func (d *DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Executor returns the standard DB executor.
func (d *DB) Executor() DBExecutor {
	return d.db
}

// Backup performs an atomic, live SQLite Online Backup using VACUUM INTO.
func (d *DB) Backup(ctx context.Context, destPath string) error {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil && filepath.Dir(destPath) != "." {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	// Remove target file if it already exists because VACUUM INTO requires a non-existent destination
	_ = os.Remove(destPath)

	cleanDest := filepath.ToSlash(destPath)
	query := fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(cleanDest, "'", "''"))
	if _, err := d.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("executing sqlite online backup: %w", err)
	}

	d.logger.Info("SQLite online backup created successfully", "destination", destPath)
	return nil
}

// WithTx executes a function inside an exclusive write transaction.
func (d *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v (rollback error: %v)", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// Migrate reads embedded SQL files and executes any unapplied migrations in version order.
func (d *DB) Migrate(ctx context.Context) error {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	// Ensure schema_migrations exists
	_, err := d.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading embedded migrations: %w", err)
	}

	type migration struct {
		version int
		name    string
	}
	var migrations []migration

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		ver, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		migrations = append(migrations, migration{version: ver, name: entry.Name()})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	for _, m := range migrations {
		var applied int
		err := d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&applied)
		if err != nil {
			return fmt.Errorf("checking migration status for version %d: %w", m.version, err)
		}
		if applied > 0 {
			continue
		}

		d.logger.Info("Applying migration", "version", m.version, "name", m.name)

		content, err := migrationFS.ReadFile(path.Join("migrations", m.name))
		if err != nil {
			return fmt.Errorf("reading migration file %q: %w", m.name, err)
		}

		tx, err := d.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning migration tx %d: %w", m.version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", m.name, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.version, err)
		}
	}

	return nil
}
