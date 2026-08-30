package http

import (
	"net/http"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type ArtistMetaHandler struct {
	metaService *service.ArtistMetaService
}

func NewArtistMetaHandler(metaService *service.ArtistMetaService) *ArtistMetaHandler {
	return &ArtistMetaHandler{metaService: metaService}
}

func (h *ArtistMetaHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{artistId}/biography", h.GetBiography)
	return r
}

func (h *ArtistMetaHandler) GetBiography(w http.ResponseWriter, r *http.Request) {
	artistID := chi.URLParam(r, "artistId")
	if artistID == "" {
		RespondError(w, r, domain.ErrInvalidInput("artistId is required"))
		return
	}

	bio, err := h.metaService.GetArtistBiography(r.Context(), artistID)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, bio)
}
