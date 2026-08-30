package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	catalogService *service.CatalogService
}

func NewAdminHandler(catalogService *service.CatalogService) *AdminHandler {
	return &AdminHandler{catalogService: catalogService}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(RequireAdmin)

	r.Get("/libraries", h.ListLibraries)
	r.Post("/libraries", h.CreateLibrary)
	r.Post("/scan", h.TriggerScan)
	r.Get("/stats", h.GetStats)
	r.Get("/scan/events", h.ScanEventsSSE)

	return r
}

type CreateLibraryRequest struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type TriggerScanRequest struct {
	LibraryID string `json:"libraryId"`
}

func (h *AdminHandler) ListLibraries(w http.ResponseWriter, r *http.Request) {
	libs, err := h.catalogService.ListLibraries(r.Context())
	if err != nil {
		RespondError(w, r, err)
		return
	}

	if libs == nil {
		libs = []*domain.Library{}
	}

	JSON(w, http.StatusOK, CollectionResponse[*domain.Library]{
		Items: libs,
		Total: len(libs),
	})
}

func (h *AdminHandler) CreateLibrary(w http.ResponseWriter, r *http.Request) {
	var req CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	lib, err := h.catalogService.CreateLibrary(r.Context(), req.Name, req.Path)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusCreated, lib)
}

func (h *AdminHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	var req TriggerScanRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.LibraryID == "" {
		req.LibraryID = r.URL.Query().Get("libraryId")
	}
	if req.LibraryID == "" {
		RespondError(w, r, domain.ErrInvalidInput("libraryId is required in body or query"))
		return
	}

	progress, err := h.catalogService.TriggerScan(r.Context(), req.LibraryID)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, progress)
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.catalogService.GetStats(r.Context())
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, stats)
}

// ScanEventsSSE provides Server-Sent Events for real-time scan progress.
func (h *AdminHandler) ScanEventsSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			libs, err := h.catalogService.ListLibraries(r.Context())
			if err == nil {
				data, _ := json.Marshal(map[string]any{
					"type":      "scan:status",
					"libraries": libs,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
