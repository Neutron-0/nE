package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete runtime configuration for nE.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Auth      AuthConfig      `yaml:"auth"`
	Scanner   ScannerConfig   `yaml:"scanner"`
	Streaming StreamingConfig `yaml:"streaming"`
	Artwork   ArtworkConfig   `yaml:"artwork"`
	Paths     PathsConfig     `yaml:"paths"`
}

type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	BasePath        string        `yaml:"base_path"`
	LogLevel        string        `yaml:"log_level"`
	LogFormat       string        `yaml:"log_format"` // "json" or "text"
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	TrustedProxies  []string      `yaml:"trusted_proxies"`
	CORSAllowAll    bool          `yaml:"cors_allow_all"`
}

type DatabaseConfig struct {
	Path        string `yaml:"path"`
	BusyTimeout int    `yaml:"busy_timeout"` // in milliseconds
}

type AuthConfig struct {
	JWTSecretKey       string        `yaml:"jwt_secret_key"`
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
	RateLimitRequests  int           `yaml:"rate_limit_requests"` // max attempts per window
	RateLimitWindowSec int           `yaml:"rate_limit_window_sec"`
}

type ScannerConfig struct {
	ScanOnStartup bool   `yaml:"scan_on_startup"`
	WorkerCount   int    `yaml:"worker_count"`
	BatchSize     int    `yaml:"batch_size"`
	IgnoreFile    string `yaml:"ignore_file"`
	CleanEnglish  bool   `yaml:"clean_english_articles"` // "The Beatles" -> "Beatles, The"
}

type StreamingConfig struct {
	MaxConcurrentTranscodes int    `yaml:"max_concurrent_transcodes"`
	MaxTranscodesPerUser    int    `yaml:"max_transcodes_per_user"`
	TranscodeCacheDir       string `yaml:"transcode_cache_dir"`
	MaxCacheSizeBytes       int64  `yaml:"max_cache_size_bytes"`
	DefaultTranscodeCodec   string `yaml:"default_transcode_codec"` // "mp3", "opus", "aac"
	DefaultTranscodeBitrate int    `yaml:"default_transcode_bitrate"`
}

type ArtworkConfig struct {
	CacheDir          string `yaml:"cache_dir"`
	MaxCacheSizeBytes int64  `yaml:"max_cache_size_bytes"`
	DefaultQuality    int    `yaml:"default_quality"`
}

type PathsConfig struct {
	ConfigDir string `yaml:"config_dir"`
	DataDir   string `yaml:"data_dir"`
	CacheDir  string `yaml:"cache_dir"`
	MusicDir  string `yaml:"music_dir"`
}

// DefaultConfig returns safe production-ready defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            4533,
			BasePath:        "",
			LogLevel:        "info",
			LogFormat:       "text",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			TrustedProxies:  []string{"127.0.0.1/32", "::1/128"},
			CORSAllowAll:    true,
		},
		Database: DatabaseConfig{
			Path:        filepath.Join("data", "ne.db"),
			BusyTimeout: 5000,
		},
		Auth: AuthConfig{
			JWTSecretKey:       "",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 30 * 24 * time.Hour,
			RateLimitRequests:  10,
			RateLimitWindowSec: 60,
		},
		Scanner: ScannerConfig{
			ScanOnStartup: true,
			WorkerCount:   4,
			BatchSize:     200,
			IgnoreFile:    ".neignore",
			CleanEnglish:  true,
		},
		Streaming: StreamingConfig{
			MaxConcurrentTranscodes: 4,
			MaxTranscodesPerUser:    2,
			TranscodeCacheDir:       filepath.Join("cache", "transcode"),
			MaxCacheSizeBytes:       10 * 1024 * 1024 * 1024, // 10 GB
			DefaultTranscodeCodec:   "mp3",
			DefaultTranscodeBitrate: 320,
		},
		Artwork: ArtworkConfig{
			CacheDir:          filepath.Join("cache", "artwork"),
			MaxCacheSizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
			DefaultQuality:    85,
		},
		Paths: PathsConfig{
			ConfigDir: "config",
			DataDir:   "data",
			CacheDir:  "cache",
			MusicDir:  "music",
		},
	}
}

// Load reads configuration from file (if provided) and applies environment variable overrides.
func Load(configFile string) (*Config, error) {
	cfg := DefaultConfig()

	if configFile != "" {
		if err := loadFromFile(configFile, cfg); err != nil {
			return nil, fmt.Errorf("loading config file %q: %w", configFile, err)
		}
	} else {
		candidates := []string{"ne.yaml", "ne.yml", "ne.json", "config/ne.yaml"}
		for _, cand := range candidates {
			if _, err := os.Stat(cand); err == nil {
				if err := loadFromFile(cand, cfg); err != nil {
					return nil, fmt.Errorf("loading discovered config %q: %w", cand, err)
				}
				break
			}
		}
	}

	// Apply environment variable overrides (NE_*)
	applyEnvOverrides(cfg)

	// Ensure essential directories exist
	if err := os.MkdirAll(cfg.Paths.DataDir, 0750); err != nil {
		return nil, fmt.Errorf("creating data directory %q: %w", cfg.Paths.DataDir, err)
	}
	if err := os.MkdirAll(cfg.Paths.CacheDir, 0750); err != nil {
		return nil, fmt.Errorf("creating cache directory %q: %w", cfg.Paths.CacheDir, err)
	}
	if err := os.MkdirAll(cfg.Paths.ConfigDir, 0750); err != nil {
		return nil, fmt.Errorf("creating config directory %q: %w", cfg.Paths.ConfigDir, err)
	}
	if err := os.MkdirAll(cfg.Artwork.CacheDir, 0750); err != nil {
		return nil, fmt.Errorf("creating artwork cache directory %q: %w", cfg.Artwork.CacheDir, err)
	}
	if err := os.MkdirAll(cfg.Streaming.TranscodeCacheDir, 0750); err != nil {
		return nil, fmt.Errorf("creating transcode cache directory %q: %w", cfg.Streaming.TranscodeCacheDir, err)
	}

	return cfg, nil
}

func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("NE_HOST"); v != "" {
		cfg.Server.Host = v
	}
	portEnv := os.Getenv("NE_PORT")
	if portEnv == "" {
		portEnv = os.Getenv("PORT")
	}
	if portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			cfg.Server.Port = p
		}
	}
	if v := os.Getenv("NE_BASE_PATH"); v != "" {
		cfg.Server.BasePath = v
	}
	if v := os.Getenv("NE_LOG_LEVEL"); v != "" {
		cfg.Server.LogLevel = strings.ToLower(v)
	}
	if v := os.Getenv("NE_LOG_FORMAT"); v != "" {
		cfg.Server.LogFormat = strings.ToLower(v)
	}
	if v := os.Getenv("NE_DB_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("NE_MUSIC_DIR"); v != "" {
		cfg.Paths.MusicDir = v
	}
	if v := os.Getenv("NE_DATA_DIR"); v != "" {
		cfg.Paths.DataDir = v
		cfg.Database.Path = filepath.Join(v, "ne.db")
	}
	if v := os.Getenv("NE_CACHE_DIR"); v != "" {
		cfg.Paths.CacheDir = v
		cfg.Artwork.CacheDir = filepath.Join(v, "artwork")
		cfg.Streaming.TranscodeCacheDir = filepath.Join(v, "transcode")
	}
	if v := os.Getenv("NE_CONFIG_DIR"); v != "" {
		cfg.Paths.ConfigDir = v
	}
	if v := os.Getenv("NE_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecretKey = v
	}
	if v := os.Getenv("NE_SCAN_ON_STARTUP"); v != "" {
		cfg.Scanner.ScanOnStartup = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("NE_MAX_CONCURRENT_TRANSCODES"); v != "" {
		if count, err := strconv.Atoi(v); err == nil {
			cfg.Streaming.MaxConcurrentTranscodes = count
		}
	}
}
