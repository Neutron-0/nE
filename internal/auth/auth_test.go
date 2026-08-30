package auth_test

import (
	"testing"
	"time"

	"ne/internal/auth"
)

func TestArgon2PasswordHashing(t *testing.T) {
	password := "SecretPassw0rd!"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	valid, err := auth.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("verification error: %v", err)
	}
	if !valid {
		t.Fatal("expected password to verify successfully")
	}

	invalid, err := auth.VerifyPassword("WrongPassword", hash)
	if err != nil {
		t.Fatalf("verification error on wrong password: %v", err)
	}
	if invalid {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestJWTManager(t *testing.T) {
	dir := t.TempDir()
	mgr, err := auth.NewJWTManager("", dir, 5*time.Minute, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWTManager: %v", err)
	}

	claims := auth.UserClaims{
		UserID:       "user-123",
		Username:     "audiodude",
		IsAdmin:      true,
		TokenVersion: 2,
	}

	tokenStr, exp, err := mgr.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if time.Until(exp) <= 0 {
		t.Fatal("expiration should be in the future")
	}

	parsed, err := mgr.VerifyAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}

	if parsed.UserID != claims.UserID {
		t.Errorf("expected userID %q, got %q", claims.UserID, parsed.UserID)
	}
	if parsed.Username != claims.Username {
		t.Errorf("expected username %q, got %q", claims.Username, parsed.Username)
	}
	if parsed.IsAdmin != claims.IsAdmin {
		t.Errorf("expected isAdmin %v, got %v", claims.IsAdmin, parsed.IsAdmin)
	}
	if parsed.TokenVersion != claims.TokenVersion {
		t.Errorf("expected tokenVersion %d, got %d", claims.TokenVersion, parsed.TokenVersion)
	}
}

func TestRefreshTokenHashing(t *testing.T) {
	tokenStr, tokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	computedHash := auth.HashRefreshToken(tokenStr)
	if computedHash != tokenHash {
		t.Fatalf("expected hash %q, got %q", tokenHash, computedHash)
	}
}
