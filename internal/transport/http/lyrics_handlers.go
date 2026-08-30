package http

import (
	"net/http"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type LyricsHandler struct {
	lyricsService *service.LyricsService
}

func NewLyricsHandler(lyricsService *service.LyricsService) *LyricsHandler {
	return &LyricsHandler{lyricsService: lyricsService}
}

func (h *LyricsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{trackId}", h.GetLyrics)
	return r
}

func (h *LyricsHandler) GetLyrics(w http.ResponseWriter, r *http.Request) {
	trackID := chi.URLParam(r, "trackId")
	if trackID == "" {
		RespondError(w, r, domain.ErrInvalidInput("trackId is required"))
		return
	}

	lyrics, err := h.lyricsService.GetLyrics(r.Context(), trackID)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, lyrics)
}
