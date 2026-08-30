package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type AnnotationHandler struct {
	annoService *service.AnnotationService
}

func NewAnnotationHandler(annoService *service.AnnotationService) *AnnotationHandler {
	return &AnnotationHandler{annoService: annoService}
}

func (h *AnnotationHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/star", h.StarItem)
	r.Post("/rating", h.RateItem)

	return r
}

type StarRequest struct {
	ItemType  string `json:"itemType"` // "track", "album", "artist"
	ItemID    string `json:"itemId"`
	IsStarred bool   `json:"isStarred"`
}

type RatingRequest struct {
	ItemType string `json:"itemType"` // "track", "album"
	ItemID   string `json:"itemId"`
	Rating   int    `json:"rating"`   // 0-5
}

type ScrobbleRequest struct {
	TrackID        string  `json:"trackId"`
	PlayerName     string  `json:"playerName"`
	DurationPlayed float64 `json:"durationPlayed"`
	Completed      bool    `json:"completed"`
}

func (h *AnnotationHandler) StarItem(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	var req StarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	if err := h.annoService.StarItem(r.Context(), claims.UserID, req.ItemType, req.ItemID, req.IsStarred); err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"status": "ok", "isStarred": req.IsStarred})
}

func (h *AnnotationHandler) RateItem(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	var req RatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	if err := h.annoService.RateItem(r.Context(), claims.UserID, req.ItemType, req.ItemID, req.Rating); err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"status": "ok", "rating": req.Rating})
}

func (h *AnnotationHandler) Scrobble(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	var req ScrobbleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	if err := h.annoService.Scrobble(r.Context(), claims.UserID, req.TrackID, req.PlayerName, req.DurationPlayed, req.Completed); err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusCreated, map[string]any{"status": "scrobbled"})
}

func (h *AnnotationHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	tracks, err := h.annoService.GetFavorites(r.Context(), claims.UserID)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	if tracks == nil {
		tracks = []*domain.Track{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Track]{
		Items: tracks,
		Total: len(tracks),
	})
}

func (h *AnnotationHandler) GetRecentHistory(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	history, err := h.annoService.GetRecentHistory(r.Context(), claims.UserID, limit)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	if history == nil {
		history = []*domain.PlaybackRecord{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.PlaybackRecord]{
		Items: history,
		Total: len(history),
	})
}
