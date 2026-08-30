package http

import (
	"net/http"
	"strconv"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type SmartPlaylistHandler struct {
	smartService *service.SmartPlaylistService
}

func NewSmartPlaylistHandler(smartService *service.SmartPlaylistService) *SmartPlaylistHandler {
	return &SmartPlaylistHandler{smartService: smartService}
}

func (h *SmartPlaylistHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListSmartPlaylists)
	r.Get("/{id}", h.GetSmartPlaylist)

	return r
}

func (h *SmartPlaylistHandler) ListSmartPlaylists(w http.ResponseWriter, r *http.Request) {
	list := h.smartService.ListAvailableSmartPlaylists()
	JSON(w, http.StatusOK, list)
}

func (h *SmartPlaylistHandler) GetSmartPlaylist(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	id := chi.URLParam(r, "id")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	tracks, err := h.smartService.GetSmartPlaylistTracks(r.Context(), claims.UserID, id, limit)
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
