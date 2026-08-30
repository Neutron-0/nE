package http

import (
	"net/http"
	"strconv"

	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type ArtworkHandler struct {
	artworkService *service.ArtworkService
}

func NewArtworkHandler(artworkService *service.ArtworkService) *ArtworkHandler {
	return &ArtworkHandler{artworkService: artworkService}
}

func (h *ArtworkHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{type}/{id}", h.GetArtwork)
	return r
}

func (h *ArtworkHandler) GetArtwork(w http.ResponseWriter, r *http.Request) {
	itemType := chi.URLParam(r, "type")
	itemID := chi.URLParam(r, "id")
	size := 300
	if s := r.URL.Query().Get("size"); s != "" {
		if val, err := strconv.Atoi(s); err == nil && val > 0 {
			size = val
		}
	}

	if err := h.artworkService.ServeArtwork(r.Context(), w, r, itemType, itemID, size); err != nil {
		RespondError(w, r, err)
		return
	}
}
