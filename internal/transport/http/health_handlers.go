package http

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"ne/internal/config"
)

type HealthHandler struct {
	cfg       *config.Config
	startTime time.Time
	dbChecker func() error
}

func NewHealthHandler(cfg *config.Config, dbChecker func() error) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startTime: time.Now(),
		dbChecker: dbChecker,
	}
}

type LivenessResponse struct {
	Status string `json:"status"`
}

type ReadinessResponse struct {
	Status    string            `json:"status"`
	Database  string            `json:"database"`
	Storage   string            `json:"storage"`
	Timestamp time.Time         `json:"timestamp"`
}

type DiagnosticsResponse struct {
	Status        string            `json:"status"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Version       string            `json:"version"`
	GoVersion     string            `json:"go_version"`
	Platform      string            `json:"platform"`
	NumCPU        int               `json:"num_cpu"`
	FFmpeg        FFmpegDiagnostic  `json:"ffmpeg"`
	Storage       StorageDiagnostic `json:"storage"`
}

type FFmpegDiagnostic struct {
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
}

type StorageDiagnostic struct {
	DataDir   string `json:"data_dir"`
	CacheDir  string `json:"cache_dir"`
	ConfigDir string `json:"config_dir"`
	MusicDir  string `json:"music_dir"`
}

// Liveness indicates if the process is running.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, LivenessResponse{Status: "ok"})
}

// Readiness indicates if the server is capable of serving requests.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if h.dbChecker != nil {
		if err := h.dbChecker(); err != nil {
			dbStatus = "unavailable"
			JSON(w, http.StatusServiceUnavailable, ReadinessResponse{
				Status:    "unready",
				Database:  dbStatus,
				Storage:   "accessible",
				Timestamp: time.Now(),
			})
			return
		}
	}

	storageStatus := "accessible"
	if _, err := os.Stat(h.cfg.Paths.DataDir); err != nil {
		storageStatus = "inaccessible"
		JSON(w, http.StatusServiceUnavailable, ReadinessResponse{
			Status:    "unready",
			Database:  dbStatus,
			Storage:   storageStatus,
			Timestamp: time.Now(),
		})
		return
	}

	JSON(w, http.StatusOK, ReadinessResponse{
		Status:    "ready",
		Database:  dbStatus,
		Storage:   storageStatus,
		Timestamp: time.Now(),
	})
}

// Diagnostics provides runtime diagnostics for administration.
func (h *HealthHandler) Diagnostics(w http.ResponseWriter, r *http.Request) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	ffmpegAvail := err == nil
	ffmpegVer := ""
	if ffmpegAvail {
		out, err := exec.Command("ffmpeg", "-version").Output()
		if err == nil && len(out) > 0 {
			lines := string(out)
			if idx := len(lines); idx > 60 {
				ffmpegVer = lines[:60]
			} else {
				ffmpegVer = lines
			}
		}
	}

	diag := DiagnosticsResponse{
		Status:        "ok",
		UptimeSeconds: int64(time.Since(h.startTime).Seconds()),
		Version:       "0.1.0-alpha",
		GoVersion:     runtime.Version(),
		Platform:      runtime.GOOS + "/" + runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		FFmpeg: FFmpegDiagnostic{
			Available: ffmpegAvail,
			Path:      ffmpegPath,
			Version:   ffmpegVer,
		},
		Storage: StorageDiagnostic{
			DataDir:   h.cfg.Paths.DataDir,
			CacheDir:  h.cfg.Paths.CacheDir,
			ConfigDir: h.cfg.Paths.ConfigDir,
			MusicDir:  h.cfg.Paths.MusicDir,
		},
	}

	JSON(w, http.StatusOK, diag)
}
