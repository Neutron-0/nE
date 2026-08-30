package http

import (
	"net/http"

	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type StreamHandler struct {
	streamService *service.StreamService
}

func NewStreamHandler(streamService *service.StreamService) *StreamHandler {
	return &StreamHandler{streamService: streamService}
}

func (h *StreamHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", h.StreamTrack)

	return r
}

func (h *StreamHandler) StreamTrack(w http.ResponseWriter, r *http.Request) {
	trackID := chi.URLParam(r, "id")
	if err := h.streamService.StreamTrack(r.Context(), w, r, trackID); err != nil {
		RespondError(w, r, err)
		return
	}
}
