package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ne/internal/auth"
	"ne/internal/domain"
	"ne/internal/repository"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

type AuthResult struct {
	AccessToken  string       `json:"accessToken"`
	ExpiresAt    time.Time    `json:"expiresAt"`
	RefreshToken string       `json:"-"`
	User         *domain.User `json:"user"`
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// SetupInitialAdmin registers the very first admin user in nE. Returns error if users already exist.
func (s *AuthService) SetupInitialAdmin(ctx context.Context, username, email, password string) (*AuthResult, error) {
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking user count: %w", err)
	}
	if count > 0 {
		return nil, domain.ErrForbidden("Initial setup is only available on an empty instance")
	}

	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return nil, domain.ErrInvalidInput("Username must be at least 3 characters long")
	}
	if len(password) < 6 {
		return nil, domain.ErrInvalidInput("Password must be at least 6 characters long")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        strings.TrimSpace(email),
		PasswordHash: hash,
		IsAdmin:      true,
		CanTranscode: true,
		TokenVersion: 1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating initial admin user: %w", err)
	}

	return s.createSession(ctx, user)
}

// Login authenticates a user and returns an access token and refresh token.
func (s *AuthService) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrUnauthorized("Invalid username or password")
	}

	valid, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, domain.ErrUnauthorized("Invalid username or password")
	}

	return s.createSession(ctx, user)
}

// Refresh rotates the refresh token and issues a new access token.
func (s *AuthService) Refresh(ctx context.Context, refreshTokenStr string) (*AuthResult, error) {
	if refreshTokenStr == "" {
		return nil, domain.ErrUnauthorized("Missing refresh token")
	}

	tokenHash := auth.HashRefreshToken(refreshTokenStr)
	tokenRecord, err := s.userRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrUnauthorized("Invalid refresh token")
	}

	if !tokenRecord.IsActive() {
		return nil, domain.ErrUnauthorized("Refresh token has expired or been revoked")
	}

	// Revoke current refresh token (rotation)
	_ = s.userRepo.RevokeRefreshToken(ctx, tokenRecord.ID)

	user, err := s.userRepo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized("User not found")
	}

	return s.createSession(ctx, user)
}

// Logout revokes the given refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	if refreshTokenStr == "" {
		return nil
	}
	tokenHash := auth.HashRefreshToken(refreshTokenStr)
	tokenRecord, err := s.userRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil
	}
	return s.userRepo.RevokeRefreshToken(ctx, tokenRecord.ID)
}

// GetCurrentUser returns the user profile for the given ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// ChangePassword updates user password and revokes all active sessions.
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return domain.ErrInvalidInput("New password must be at least 6 characters long")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	valid, err := auth.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !valid {
		return domain.ErrUnauthorized("Incorrect current password")
	}

	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}

	// Invalidate all existing refresh tokens
	return s.userRepo.RevokeAllUserRefreshTokens(ctx, userID)
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*AuthResult, error) {
	claims := auth.UserClaims{
		UserID:       user.ID,
		Username:     user.Username,
		IsAdmin:      user.IsAdmin,
		TokenVersion: user.TokenVersion,
	}

	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshTokenStr, tokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
	}

	if err := s.userRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("persisting refresh token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshTokenStr,
		User:         user,
	}, nil
}
