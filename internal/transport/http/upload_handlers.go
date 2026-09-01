package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ne/internal/domain"
	"ne/internal/service"
)

type UploadHandler struct {
	catalogService *service.CatalogService
	githubSync     *service.GitHubSyncService
}

func NewUploadHandler(catalogService *service.CatalogService, githubSync *service.GitHubSyncService) *UploadHandler {
	return &UploadHandler{
		catalogService: catalogService,
		githubSync:     githubSync,
	}
}

type UploadResponse struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Files     []string `json:"files"`
	LibraryID string   `json:"libraryId"`
}

func (h *UploadHandler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form up to 100 MB memory buffer
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Failed to parse multipart form or file exceeds size limit"))
		return
	}

	form := r.MultipartForm
	files := form.File["files"]
	if len(files) == 0 {
		if single := form.File["file"]; len(single) > 0 {
			files = single
		}
	}

	if len(files) == 0 {
		RespondError(w, r, domain.ErrInvalidInput("No audio files provided in 'files' or 'file' form field"))
		return
	}

	// Determine destination library
	libs, err := h.catalogService.ListLibraries(r.Context())
	if err != nil {
		RespondError(w, r, err)
		return
	}

	var targetLib *domain.Library
	reqLibID := r.FormValue("libraryId")
	if reqLibID != "" {
		for _, lib := range libs {
			if lib.ID == reqLibID {
				targetLib = lib
				break
			}
		}
	}

	if targetLib == nil {
		if len(libs) > 0 {
			targetLib = libs[0]
		} else {
			// Auto-create default library directory if none exists
			defaultMusicDir := "./music"
			if os.Getenv("MUSIC_PATH") != "" {
				defaultMusicDir = os.Getenv("MUSIC_PATH")
			}
			_ = os.MkdirAll(defaultMusicDir, 0755)
			newLib, createErr := h.catalogService.CreateLibrary(r.Context(), "Default Library", defaultMusicDir)
			if createErr != nil {
				RespondError(w, r, fmt.Errorf("failed to auto-create default library: %w", createErr))
				return
			}
			targetLib = newLib
		}
	}

	// Supported audio formats
	allowedExts := map[string]bool{
		".mp3":  true,
		".flac": true,
		".m4a":  true,
		".ogg":  true,
		".wav":  true,
		".aac":  true,
		".opus": true,
		".wma":  true,
		".aiff": true,
	}

	var savedFiles []string
	for _, fileHeader := range files {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !allowedExts[ext] {
			continue // Skip non-audio files
		}

		cleanName := filepath.Base(fileHeader.Filename)
		destPath := filepath.Join(targetLib.Path, cleanName)

		// Prevent arbitrary file overwrite: if file exists, add timestamp suffix
		if _, err := os.Stat(destPath); err == nil {
			cleanName = fmt.Sprintf("%s_%d%s", strings.TrimSuffix(cleanName, ext), time.Now().UnixNano(), ext)
			destPath = filepath.Join(targetLib.Path, cleanName)
		}

		// Process single file copy with immediate file descriptor closure
		saveErr := func() error {
			src, err := fileHeader.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			return err
		}()

		if saveErr == nil {
			savedFiles = append(savedFiles, cleanName)
		}
	}

	if len(savedFiles) == 0 {
		RespondError(w, r, domain.ErrInvalidInput("No valid supported audio files were uploaded"))
		return
	}

	// Trigger library scan asynchronously to ingest metadata and update database
	go func(libID string) {
		_, _ = h.catalogService.TriggerScan(context.Background(), libID)
	}(targetLib.ID)

	// Trigger GitHub repository backup asynchronously if configured
	if h.githubSync != nil {
		go func(libPath string, files []string) {
			for _, f := range files {
				fullPath := filepath.Join(libPath, f)
				_, _ = h.githubSync.BackupTrack(context.Background(), fullPath, f)
			}
		}(targetLib.Path, savedFiles)
	}

	JSON(w, http.StatusOK, UploadResponse{
		Success:   true,
		Message:   fmt.Sprintf("Successfully saved %d audio file(s). Automatic database ingestion & GitHub backup initiated.", len(savedFiles)),
		Files:     savedFiles,
		LibraryID: targetLib.ID,
	})
}
