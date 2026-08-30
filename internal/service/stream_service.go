package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ne/internal/domain"
	"ne/internal/media"
	"ne/internal/repository"
)

type StreamService struct {
	catalogRepo *repository.CatalogRepository
	transcoder  *media.Transcoder
}

func NewStreamService(catalogRepo *repository.CatalogRepository, transcoder *media.Transcoder) *StreamService {
	if transcoder == nil {
		transcoder = media.NewTranscoder(4)
	}
	return &StreamService{
		catalogRepo: catalogRepo,
		transcoder:  transcoder,
	}
}

// StreamTrack streams an audio file directly with RFC 7233 Range support or transcodes via FFmpeg if requested.
func (s *StreamService) StreamTrack(ctx context.Context, w http.ResponseWriter, r *http.Request, trackID string) error {
	track, err := s.catalogRepo.GetTrackByID(ctx, trackID)
	if err != nil {
		return err
	}

	lib, err := s.catalogRepo.GetLibraryByID(ctx, track.LibraryID)
	if err != nil {
		return err
	}

	// Security Jail Containment Check
	if err := verifyPathContainment(lib.Path, track.Path); err != nil {
		return domain.ErrForbidden("Access denied: File outside library root")
	}

	// Check if transcoding is explicitly requested
	formatReq := strings.ToLower(r.URL.Query().Get("format"))
	bitrateStr := r.URL.Query().Get("bitrate")

	if formatReq != "" && s.transcoder.IsAvailable() && formatReq != strings.ToLower(track.Format) {
		bitrate := 192
		if val, err := strconv.Atoi(bitrateStr); err == nil && val > 0 {
			bitrate = val
		}
		var offset float64
		if offStr := r.URL.Query().Get("offset"); offStr != "" {
			if val, err := strconv.ParseFloat(offStr, 64); err == nil {
				offset = val
			}
		}

		w.Header().Set("Content-Type", getAudioContentType(formatReq))
		w.Header().Set("Transfer-Encoding", "chunked")

		return s.transcoder.Stream(r.Context(), w, media.TranscodeOptions{
			SourcePath:    track.Path,
			Format:        formatReq,
			BitrateKbps:   bitrate,
			OffsetSeconds: offset,
		})
	}

	// Direct Play with standard RFC 7233 Range support
	file, err := os.Open(track.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.ErrNotFound("Audio file", track.Path)
		}
		return domain.NewAppError(domain.ErrCodeStorageUnavailable, "Music storage currently inaccessible", http.StatusServiceUnavailable, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return domain.NewAppError(domain.ErrCodeStorageUnavailable, "Cannot inspect audio file", http.StatusServiceUnavailable, err)
	}

	w.Header().Set("Content-Type", getAudioContentType(track.Format))
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, track.Filename, stat.ModTime(), file)
	return nil
}

func verifyPathContainment(libraryRoot, targetPath string) error {
	cleanTarget := filepath.Clean(targetPath)
	cleanRoot := filepath.Clean(libraryRoot)

	realTarget, err := filepath.EvalSymlinks(cleanTarget)
	if err != nil {
		realTarget = cleanTarget
	}

	realRoot, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil {
		realRoot = cleanRoot
	}

	rel, err := filepath.Rel(realRoot, realTarget)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("path traversal attempt")
	}

	return nil
}

func getAudioContentType(format string) string {
	switch strings.ToLower(format) {
	case "flac":
		return "audio/flac"
	case "mp3":
		return "audio/mpeg"
	case "m4a", "mp4", "aac":
		return "audio/mp4"
	case "ogg", "oga":
		return "audio/ogg"
	case "opus":
		return "audio/ogg; codecs=opus"
	case "wav":
		return "audio/wav"
	case "aiff", "aif":
		return "audio/aiff"
	default:
		return "application/octet-stream"
	}
}
