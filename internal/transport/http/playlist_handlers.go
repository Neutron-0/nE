package http

import (
	"encoding/json"
	"net/http"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type PlaylistHandler struct {
	playlistService *service.PlaylistService
}

func NewPlaylistHandler(playlistService *service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlistService: playlistService}
}

func (h *PlaylistHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListPlaylists)
	r.Post("/", h.CreatePlaylist)
	r.Get("/{id}", h.GetPlaylist)
	r.Put("/{id}", h.UpdatePlaylist)
	r.Put("/{id}/tracks", h.SetPlaylistTracks)
	r.Delete("/{id}", h.DeletePlaylist)

	return r
}

type CreatePlaylistRequest struct {
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	IsPublic bool   `json:"isPublic"`
}

type SetPlaylistTracksRequest struct {
	TrackIDs []string `json:"trackIds"`
}

func (h *PlaylistHandler) ListPlaylists(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	playlists, err := h.playlistService.ListPlaylists(r.Context(), claims.UserID)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	if playlists == nil {
		playlists = []*domain.Playlist{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Playlist]{
		Items: playlists,
		Total: len(playlists),
	})
}

func (h *PlaylistHandler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	var req CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	p, err := h.playlistService.CreatePlaylist(r.Context(), claims.UserID, req.Name, req.Comment, req.IsPublic)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusCreated, p)
}

func (h *PlaylistHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.playlistService.GetPlaylist(r.Context(), id)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, p)
}

func (h *PlaylistHandler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	id := chi.URLParam(r, "id")
	var req CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	p, err := h.playlistService.UpdatePlaylist(r.Context(), claims.UserID, id, req.Name, req.Comment, req.IsPublic)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, p)
}

func (h *PlaylistHandler) SetPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	id := chi.URLParam(r, "id")
	var req SetPlaylistTracksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	if err := h.playlistService.SetPlaylistTracks(r.Context(), claims.UserID, id, req.TrackIDs); err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PlaylistHandler) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	claims := GetAuthClaims(r.Context())
	if claims == nil {
		RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.playlistService.DeletePlaylist(r.Context(), claims.UserID, id); err != nil {
		RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
