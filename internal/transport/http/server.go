package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ne/internal/auth"
	"ne/internal/config"
	"ne/internal/domain"
	"ne/internal/service"
	"ne/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server wraps chi.Router and the underlying http.Server.
type Server struct {
	cfg         *config.Config
	logger      *slog.Logger
	router      chi.Router
	server      *http.Server
	healthH     *HealthHandler
	authH       *AuthHandler
	catalogH    *CatalogHandler
	adminH      *AdminHandler
	streamH     *StreamHandler
	artworkH    *ArtworkHandler
	annoH       *AnnotationHandler
	playlistH   *PlaylistHandler
	jwtManager  *auth.JWTManager
}

func NewServer(
	cfg *config.Config,
	logger *slog.Logger,
	dbChecker func() error,
	jwtManager *auth.JWTManager,
	authService *service.AuthService,
	catalogService *service.CatalogService,
	streamService *service.StreamService,
	artworkService *service.ArtworkService,
	annoService *service.AnnotationService,
	playlistService *service.PlaylistService,
) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg:        cfg,
		logger:     logger,
		router:     r,
		healthH:    NewHealthHandler(cfg, dbChecker),
		jwtManager: jwtManager,
	}

	if authService != nil {
		s.authH = NewAuthHandler(authService)
	}
	if catalogService != nil {
		s.catalogH = NewCatalogHandler(catalogService)
		s.adminH = NewAdminHandler(catalogService)
	}
	if streamService != nil {
		s.streamH = NewStreamHandler(streamService)
	}
	if artworkService != nil {
		s.artworkH = NewArtworkHandler(artworkService)
	}
	if annoService != nil {
		s.annoH = NewAnnotationHandler(annoService)
	}
	if playlistService != nil {
		s.playlistH = NewPlaylistHandler(playlistService)
	}

	s.setupMiddlewares()
	s.setupRoutes()

	return s
}

func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) setupMiddlewares() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(SlogRequestLogger(s.logger))
	s.router.Use(middleware.Recoverer)
	s.router.Use(SecurityHeadersMiddleware())
	s.router.Use(CORSMiddleware(s.cfg.Server.CORSAllowAll))
	s.router.Use(middleware.Compress(5))

	if s.jwtManager != nil {
		s.router.Use(AuthMiddleware(s.jwtManager))
	}
}

func (s *Server) setupRoutes() {
	// Base API router
	s.router.Route("/api/v1", func(r chi.Router) {
		// Health & Diagnostics
		r.Route("/health", func(r chi.Router) {
			r.Get("/liveness", s.healthH.Liveness)
			r.Get("/readiness", s.healthH.Readiness)
			r.Get("/diagnostics", s.healthH.Diagnostics)
		})

		// Auth Endpoints
		if s.authH != nil {
			r.Mount("/auth", s.authH.Routes())
		}

		// Artwork Endpoints (Public/Protected with Token)
		if s.artworkH != nil {
			r.Mount("/artwork", s.artworkH.Routes())
		}

		// Catalog Endpoints (Protected)
		if s.catalogH != nil {
			r.Group(func(cr chi.Router) {
				cr.Use(RequireAuth)
				cr.Mount("/", s.catalogH.Routes())
			})
		}

		// Audio Streaming Endpoints (Protected via Header or Query JWT)
		if s.streamH != nil {
			r.Group(func(sr chi.Router) {
				sr.Use(RequireAuth)
				sr.Mount("/stream", s.streamH.Routes())
			})
		}

		// Annotations & Scrobble Endpoints (Protected)
		if s.annoH != nil {
			r.Group(func(ar chi.Router) {
				ar.Use(RequireAuth)
				ar.Mount("/annotations", s.annoH.Routes())
				ar.Post("/playback/scrobble", s.annoH.Scrobble)
				ar.Get("/favorites", s.annoH.GetFavorites)
				ar.Get("/history/recent", s.annoH.GetRecentHistory)
			})
		}

		// Playlist Endpoints (Protected)
		if s.playlistH != nil {
			r.Group(func(pr chi.Router) {
				pr.Use(RequireAuth)
				pr.Mount("/playlists", s.playlistH.Routes())
			})
		}

		// Admin Endpoints (Admin Only)
		if s.adminH != nil {
			r.Mount("/admin", s.adminH.Routes())
		}
	})

	// Serve Embedded React Web Application for all non-API paths
	spaHandler := web.SPAHandler()
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			RespondError(w, r, domain.ErrNotFound("Route", r.URL.Path))
			return
		}
		spaHandler.ServeHTTP(w, r)
	})
}

// MountAPI attaches sub-routers under /api/v1.
func (s *Server) MountAPI(pattern string, handler http.Handler) {
	s.router.Mount("/api/v1"+pattern, handler)
}

// Start runs the HTTP server listening on the configured host and port.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       s.cfg.Server.ReadTimeout,
		WriteTimeout:      s.cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("Starting nE HTTP Server", "address", addr, "port", s.cfg.Server.Port)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("HTTP server listen failed: %w", err)
	case <-ctx.Done():
		s.logger.Info("Shutting down HTTP server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
}
