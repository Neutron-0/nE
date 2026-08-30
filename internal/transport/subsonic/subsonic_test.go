package subsonic_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"ne/internal/auth"
	"ne/internal/domain"
	"ne/internal/repository"
	"ne/internal/service"
	"ne/internal/transport/subsonic"
)

func TestSubsonicEndpoints(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subsonic_test.db")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := repository.Open(dbPath, 5000, logger)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCatalogRepository(db)
	annoRepo := repository.NewAnnotationRepository(db)

	// Create test user: neo / neo03
	passHash, _ := auth.HashPassword("neo03")
	testUser := &domain.User{
		Username:     "neo",
		Email:        "neo@ne.audio",
		PasswordHash: passHash,
		IsAdmin:      true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	catService := service.NewCatalogService(catRepo, nil)
	annoService := service.NewAnnotationService(annoRepo)
	subHandler := subsonic.NewSubsonicHandler(userRepo, catService, nil, nil, annoService)
	router := subHandler.Routes()

	// 1. Test ping.view with valid credentials
	t.Run("PingWithValidCredentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping.view?u=neo&p=neo03&v=1.16.1&c=test&f=json", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var res map[string]subsonic.Response
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}

		subRes := res["subsonic-response"]
		if subRes.Status != "ok" || subRes.Version != "1.16.1" {
			t.Errorf("unexpected ping response: status=%q, version=%q", subRes.Status, subRes.Version)
		}
	})

	// 2. Test ping.view with invalid credentials -> 40 error
	t.Run("PingWithInvalidCredentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping.view?u=neo&p=wrongpass&v=1.16.1&c=test&f=json", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var res map[string]subsonic.Response
		_ = json.NewDecoder(rec.Body).Decode(&res)
		subRes := res["subsonic-response"]

		if subRes.Status != "failed" || subRes.Error == nil || subRes.Error.Code != 40 {
			t.Errorf("expected failed status with code 40, got %+v", subRes)
		}
	})

	// 3. Test getLicense.view
	t.Run("GetLicense", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/getLicense.view?u=neo&p=neo03&v=1.16.1&c=test&f=json", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var res map[string]subsonic.Response
		_ = json.NewDecoder(rec.Body).Decode(&res)
		subRes := res["subsonic-response"]

		if subRes.License == nil || !subRes.License.Valid {
			t.Errorf("expected valid license, got %+v", subRes.License)
		}
	})

	// 4. Test getMusicFolders.view
	t.Run("GetMusicFolders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/getMusicFolders.view?u=neo&p=neo03&v=1.16.1&c=test&f=json", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var res map[string]subsonic.Response
		_ = json.NewDecoder(rec.Body).Decode(&res)
		subRes := res["subsonic-response"]

		if subRes.MusicFolders == nil || len(subRes.MusicFolders.Folder) == 0 {
			t.Errorf("expected music folders list, got %+v", subRes.MusicFolders)
		}
	})
}
