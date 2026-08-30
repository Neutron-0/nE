package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"ne/internal/domain"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	query := `INSERT INTO users (id, username, email, password_hash, is_admin, can_transcode, token_version, subsonic_salt, subsonic_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.db.Executor().ExecContext(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash, u.IsAdmin, u.CanTranscode, u.TokenVersion, u.SubsonicSalt, u.SubsonicToken, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, is_admin, can_transcode, token_version, subsonic_salt, subsonic_token, created_at, updated_at
		FROM users WHERE id = ?`
	var u domain.User
	var email, salt, token sql.NullString

	err := r.db.Executor().QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &email, &u.PasswordHash, &u.IsAdmin, &u.CanTranscode, &u.TokenVersion, &salt, &token, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("User", id)
		}
		return nil, fmt.Errorf("getting user by id: %w", err)
	}

	if email.Valid {
		u.Email = email.String
	}
	if salt.Valid {
		u.SubsonicSalt = salt.String
	}
	if token.Valid {
		u.SubsonicToken = token.String
	}

	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, is_admin, can_transcode, token_version, subsonic_salt, subsonic_token, created_at, updated_at
		FROM users WHERE username = ? COLLATE NOCASE`
	var u domain.User
	var email, salt, token sql.NullString

	err := r.db.Executor().QueryRowContext(ctx, query, username).Scan(
		&u.ID, &u.Username, &email, &u.PasswordHash, &u.IsAdmin, &u.CanTranscode, &u.TokenVersion, &salt, &token, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("User", username)
		}
		return nil, fmt.Errorf("getting user by username: %w", err)
	}

	if email.Valid {
		u.Email = email.String
	}
	if salt.Valid {
		u.SubsonicSalt = salt.String
	}
	if token.Valid {
		u.SubsonicToken = token.String
	}

	return &u, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Executor().QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Executor().ExecContext(ctx, query, passwordHash, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("updating user password: %w", err)
	}
	return nil
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`
	t.CreatedAt = time.Now().UTC()
	_, err := r.db.Executor().ExecContext(ctx, query, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating refresh token: %w", err)
	}
	return nil
}

func (r *UserRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, created_at, revoked_at FROM refresh_tokens WHERE token_hash = ?`
	var t domain.RefreshToken
	var revokedAt sql.NullTime

	err := r.db.Executor().QueryRowContext(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &revokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound("RefreshToken", tokenHash)
		}
		return nil, fmt.Errorf("getting refresh token: %w", err)
	}

	if revokedAt.Valid {
		t.RevokedAt = &revokedAt.Time
	}

	return &t, nil
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.db.Executor().ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE id = ?`, now, id)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}

func (r *UserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.db.Executor().ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`, now, userID)
	if err != nil {
		return fmt.Errorf("revoking user refresh tokens: %w", err)
	}
	return nil
}
