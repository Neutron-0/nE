package http

import (
	"net/http"
	"strconv"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	catalogService *service.CatalogService
}

func NewCatalogHandler(catalogService *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

func (h *CatalogHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/artists", h.ListArtists)
	r.Get("/artists/{id}", h.GetArtist)
	r.Get("/albums", h.ListAlbums)
	r.Get("/albums/{id}", h.GetAlbum)
	r.Get("/tracks", h.ListTracks)
	r.Get("/tracks/{id}", h.GetTrack)
	r.Get("/genres", h.ListGenres)
	r.Get("/search", h.Search)

	return r
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}
	return
}

func (h *CatalogHandler) ListArtists(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	artists, total, err := h.catalogService.ListArtists(r.Context(), limit, offset)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	if artists == nil {
		artists = []*domain.Artist{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Artist]{
		Items: artists,
		Total: total,
	})
}

func (h *CatalogHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	artist, err := h.catalogService.GetArtist(r.Context(), id)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, artist)
}

func (h *CatalogHandler) ListAlbums(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	albums, total, err := h.catalogService.ListAlbums(r.Context(), limit, offset)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	if albums == nil {
		albums = []*domain.Album{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Album]{
		Items: albums,
		Total: total,
	})
}

func (h *CatalogHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	album, err := h.catalogService.GetAlbum(r.Context(), id)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, album)
}

func (h *CatalogHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	tracks, total, err := h.catalogService.ListTracks(r.Context(), limit, offset)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	if tracks == nil {
		tracks = []*domain.Track{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Track]{
		Items: tracks,
		Total: total,
	})
}

func (h *CatalogHandler) GetTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	track, err := h.catalogService.GetTrack(r.Context(), id)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, track)
}

func (h *CatalogHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.catalogService.ListGenres(r.Context())
	if err != nil {
		RespondError(w, r, err)
		return
	}

	if genres == nil {
		genres = []*domain.Genre{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Genre]{
		Items: genres,
		Total: len(genres),
	})
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	res, err := h.catalogService.Search(r.Context(), q, limit)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, res)
}
