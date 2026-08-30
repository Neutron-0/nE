package service_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"ne/internal/auth"
	"ne/internal/repository"
	"ne/internal/service"
)

func setupTestAuthService(t *testing.T) (*service.AuthService, *repository.DB) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "auth_test.db")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := repository.Open(dbPath, 5000, logger)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userRepo := repository.NewUserRepository(db)
	jwtMgr, err := auth.NewJWTManager("test-secret-key-at-least-32-chars-long", dir, 15*time.Minute, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	svc := service.NewAuthService(userRepo, jwtMgr)
	return svc, db
}

func TestAuthService_FullLifecycle(t *testing.T) {
	svc, _ := setupTestAuthService(t)
	ctx := context.Background()

	// 1. Initial Setup
	res, err := svc.SetupInitialAdmin(ctx, "admin", "admin@ne.audio", "adminpassword123")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if res.User.Username != "admin" || !res.User.IsAdmin {
		t.Fatalf("expected admin user, got %+v", res.User)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("expected both access and refresh tokens")
	}

	// 2. Prevent Duplicate Setup
	_, err = svc.SetupInitialAdmin(ctx, "hacker", "hacker@ne.audio", "hackerpassword")
	if err == nil {
		t.Fatal("expected second setup to be rejected")
	}

	// 3. Login
	loginRes, err := svc.Login(ctx, "admin", "adminpassword123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginRes.AccessToken == "" {
		t.Fatal("login did not return access token")
	}

	// 4. Failed Login on Bad Password
	_, err = svc.Login(ctx, "admin", "wrongpassword")
	if err == nil {
		t.Fatal("expected login to fail on wrong password")
	}

	// 5. Refresh Session & Token Rotation
	refreshed, err := svc.Refresh(ctx, loginRes.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("refreshed session missing tokens")
	}

	// 6. Old Refresh Token Re-use should fail (Rotation security)
	_, err = svc.Refresh(ctx, loginRes.RefreshToken)
	if err == nil {
		t.Fatal("expected rotated refresh token to be rejected on reuse")
	}

	// 7. Change Password
	err = svc.ChangePassword(ctx, res.User.ID, "adminpassword123", "newpassword999")
	if err != nil {
		t.Fatalf("change password failed: %v", err)
	}

	// Old password login should fail
	_, err = svc.Login(ctx, "admin", "adminpassword123")
	if err == nil {
		t.Fatal("login with old password should fail")
	}

	// New password login should succeed
	newLogin, err := svc.Login(ctx, "admin", "newpassword999")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	if newLogin.User.TokenVersion <= res.User.TokenVersion {
		t.Errorf("token version should have incremented, got %d", newLogin.User.TokenVersion)
	}
}
