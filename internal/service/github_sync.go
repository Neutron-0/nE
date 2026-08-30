package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ne/internal/config"
	"ne/internal/repository"
)

type GitHubSyncStatus struct {
	Configured bool   `json:"configured"`
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	AutoSync   bool   `json:"autoSync"`
	LastSyncAt string `json:"lastSyncAt,omitempty"`
	LastError  string `json:"lastError,omitempty"`
	TokenHint  string `json:"tokenHint,omitempty"`
}

type GitHubSyncService struct {
	db     *repository.DB
	cfg    *config.GitHubConfig
	logger *slog.Logger
	client *http.Client
	mu     sync.RWMutex

	lastSyncAt time.Time
	lastError  string
}

func NewGitHubSyncService(db *repository.DB, cfg *config.GitHubConfig, logger *slog.Logger) *GitHubSyncService {
	if logger == nil {
		logger = slog.Default()
	}
	s := &GitHubSyncService{
		db:     db,
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: 60 * time.Second},
	}
	s.loadFromDB()
	return s
}

func (s *GitHubSyncService) loadFromDB() {
	if s.db == nil {
		return
	}
	var token, repo, branch, autoSync string
	_ = s.db.Executor().QueryRowContext(context.Background(), "SELECT value FROM system_settings WHERE key = 'github_token'").Scan(&token)
	_ = s.db.Executor().QueryRowContext(context.Background(), "SELECT value FROM system_settings WHERE key = 'github_repo'").Scan(&repo)
	_ = s.db.Executor().QueryRowContext(context.Background(), "SELECT value FROM system_settings WHERE key = 'github_branch'").Scan(&branch)
	_ = s.db.Executor().QueryRowContext(context.Background(), "SELECT value FROM system_settings WHERE key = 'github_auto_sync'").Scan(&autoSync)

	s.mu.Lock()
	defer s.mu.Unlock()

	if token != "" {
		s.cfg.Token = token
	}
	if repo != "" {
		s.cfg.Repo = repo
	}
	if branch != "" {
		s.cfg.Branch = branch
	}
	if autoSync != "" {
		s.cfg.AutoSync = autoSync == "true" || autoSync == "1"
	}
}

func (s *GitHubSyncService) GetStatus() GitHubSyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hint := ""
	if len(s.cfg.Token) > 6 {
		hint = s.cfg.Token[:4] + "••••" + s.cfg.Token[len(s.cfg.Token)-2:]
	} else if s.cfg.Token != "" {
		hint = "••••••••"
	}

	lastSyncStr := ""
	if !s.lastSyncAt.IsZero() {
		lastSyncStr = s.lastSyncAt.Format(time.RFC3339)
	}

	return GitHubSyncStatus{
		Configured: s.cfg.Token != "" && s.cfg.Repo != "",
		Repo:       s.cfg.Repo,
		Branch:     s.cfg.Branch,
		AutoSync:   s.cfg.AutoSync,
		LastSyncAt: lastSyncStr,
		LastError:  s.lastError,
		TokenHint:  hint,
	}
}

func (s *GitHubSyncService) UpdateConfig(ctx context.Context, token, repo, branch string, autoSync bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if token != "" {
		s.cfg.Token = strings.TrimSpace(token)
	}
	if repo != "" {
		s.cfg.Repo = strings.TrimSpace(repo)
	}
	if branch != "" {
		s.cfg.Branch = strings.TrimSpace(branch)
	}
	s.cfg.AutoSync = autoSync

	if s.db != nil {
		queries := []struct {
			k, v string
		}{
			{"github_token", s.cfg.Token},
			{"github_repo", s.cfg.Repo},
			{"github_branch", s.cfg.Branch},
			{"github_auto_sync", fmt.Sprintf("%v", s.cfg.AutoSync)},
		}
		for _, q := range queries {
			_, err := s.db.Executor().ExecContext(ctx,
				"INSERT INTO system_settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP",
				q.k, q.v,
			)
			if err != nil {
				return fmt.Errorf("saving github setting %s: %w", q.k, err)
			}
		}
	}

	return nil
}

func (s *GitHubSyncService) TestConnection(ctx context.Context) error {
	s.mu.RLock()
	token := s.cfg.Token
	repo := s.cfg.Repo
	s.mu.RUnlock()

	if token == "" {
		return errors.New("GitHub token is not configured")
	}
	if repo == "" {
		return errors.New("GitHub repository is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/repos/%s", repo), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("connecting to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type gitHubContentResponse struct {
	SHA string `json:"sha"`
}

type gitHubCommitResponse struct {
	Content struct {
		HTMLURL string `json:"html_url"`
	} `json:"content"`
	Commit struct {
		SHA     string `json:"sha"`
		HTMLURL string `json:"html_url"`
	} `json:"commit"`
}

// BackupFile commits a local file directly to the GitHub repository.
func (s *GitHubSyncService) BackupFile(ctx context.Context, localFilePath, repoPath, commitMsg string) (string, error) {
	s.mu.RLock()
	token := s.cfg.Token
	repo := s.cfg.Repo
	branch := s.cfg.Branch
	s.mu.RUnlock()

	if token == "" || repo == "" {
		return "", errors.New("GitHub sync is not configured (missing token or repository)")
	}
	if branch == "" {
		branch = "main"
	}

	data, err := os.ReadFile(localFilePath)
	if err != nil {
		return "", fmt.Errorf("reading file %q: %w", localFilePath, err)
	}

	// 1. Check if the file already exists in GitHub to retrieve its SHA
	getURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", repo, repoPath, branch)
	getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
	if err != nil {
		return "", err
	}
	getReq.Header.Set("Authorization", "Bearer "+token)
	getReq.Header.Set("Accept", "application/vnd.github+json")
	getReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	var existingSHA string
	getResp, err := s.client.Do(getReq)
	if err == nil {
		if getResp.StatusCode == http.StatusOK {
			var fileInfo gitHubContentResponse
			if err := json.NewDecoder(getResp.Body).Decode(&fileInfo); err == nil {
				existingSHA = fileInfo.SHA
			}
		}
		_ = getResp.Body.Close()
	}

	// 2. Base64 encode the payload
	encoded := base64.StdEncoding.EncodeToString(data)

	payload := map[string]any{
		"message": commitMsg,
		"content": encoded,
		"branch":  branch,
	}
	if existingSHA != "" {
		payload["sha"] = existingSHA
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	putURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, repoPath)
	putReq, err := http.NewRequestWithContext(ctx, "PUT", putURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	putReq.Header.Set("Authorization", "Bearer "+token)
	putReq.Header.Set("Accept", "application/vnd.github+json")
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	putResp, err := s.client.Do(putReq)
	if err != nil {
		s.recordError(err.Error())
		return "", fmt.Errorf("uploading to GitHub: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(putResp.Body)
		errMsg := fmt.Sprintf("GitHub API returned %d: %s", putResp.StatusCode, string(respBody))
		s.recordError(errMsg)
		return "", errors.New(errMsg)
	}

	var commitRes gitHubCommitResponse
	_ = json.NewDecoder(putResp.Body).Decode(&commitRes)

	s.mu.Lock()
	s.lastSyncAt = time.Now()
	s.lastError = ""
	s.mu.Unlock()

	s.logger.Info("Successfully backed up file to GitHub", "path", repoPath, "commit", commitRes.Commit.SHA)
	return commitRes.Commit.SHA, nil
}

func (s *GitHubSyncService) BackupTrack(ctx context.Context, localFilePath, fileName string) (string, error) {
	s.mu.RLock()
	autoSync := s.cfg.AutoSync
	configured := s.cfg.Token != "" && s.cfg.Repo != ""
	s.mu.RUnlock()

	if !configured || !autoSync {
		return "", nil // Not configured or auto-sync disabled
	}

	repoPath := "music/" + fileName
	commitMsg := fmt.Sprintf("feat(music): add %s to library [backup via nE UI]", fileName)
	return s.BackupFile(ctx, localFilePath, repoPath, commitMsg)
}

func (s *GitHubSyncService) BackupDatabase(ctx context.Context, dbPath string) (string, error) {
	commitMsg := fmt.Sprintf("chore(db): snapshot database state at %s [backup via nE UI]", time.Now().UTC().Format(time.RFC3339))
	return s.BackupFile(ctx, dbPath, "data/ne.db", commitMsg)
}

func (s *GitHubSyncService) BackupAll(ctx context.Context, musicDir, dbPath string) (int, error) {
	s.mu.RLock()
	configured := s.cfg.Token != "" && s.cfg.Repo != ""
	s.mu.RUnlock()

	if !configured {
		return 0, errors.New("GitHub sync is not configured (missing token or repository)")
	}

	entries, err := os.ReadDir(musicDir)
	if err != nil {
		return 0, fmt.Errorf("reading music directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(musicDir, entry.Name())
		_, err := s.BackupFile(ctx, fullPath, "music/"+entry.Name(), fmt.Sprintf("feat(music): backup %s via nE UI", entry.Name()))
		if err == nil {
			count++
		} else {
			s.logger.Warn("Failed backing up track to GitHub", "file", entry.Name(), "error", err)
		}
	}

	// Also backup database if it exists
	if _, err := os.Stat(dbPath); err == nil {
		_, dbErr := s.BackupDatabase(ctx, dbPath)
		if dbErr != nil {
			s.logger.Warn("Failed backing up db to GitHub", "error", dbErr)
		}
	}

	return count, nil
}

func (s *GitHubSyncService) recordError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = msg
}
