package media

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LyricLine struct {
	Time float64 `json:"time"` // Timestamp in seconds
	Text string  `json:"text"`
}

type LyricsResult struct {
	TrackID     string      `json:"trackId"`
	IsSynced    bool        `json:"isSynced"`
	Lines       []LyricLine `json:"lines"`
	PlainLyrics string      `json:"plainLyrics,omitempty"`
	Source      string      `json:"source"` // "local", "lrclib", "none"
}

var lrcTimeRegex = regexp.MustCompile(`^\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?\](.*)$`)

// ParseLRC parses standard LRC timestamped lines into structured LyricLines.
func ParseLRC(content string) []LyricLine {
	var lines []LyricLine
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		matches := lrcTimeRegex.FindStringSubmatch(line)
		if len(matches) >= 5 {
			min, _ := strconv.Atoi(matches[1])
			sec, _ := strconv.Atoi(matches[2])
			fracStr := matches[3]
			var frac float64
			if fracStr != "" {
				if len(fracStr) == 2 {
					val, _ := strconv.Atoi(fracStr)
					frac = float64(val) / 100.0
				} else if len(fracStr) == 3 {
					val, _ := strconv.Atoi(fracStr)
					frac = float64(val) / 1000.0
				}
			}
			timestamp := float64(min*60+sec) + frac
			text := strings.TrimSpace(matches[4])
			lines = append(lines, LyricLine{
				Time: timestamp,
				Text: text,
			})
		}
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Time < lines[j].Time
	})

	return lines
}

// FindLocalLRC looks for an .lrc file in the same directory as the track.
func FindLocalLRC(trackPath string) (string, error) {
	ext := filepath.Ext(trackPath)
	lrcPath := strings.TrimSuffix(trackPath, ext) + ".lrc"

	if data, err := os.ReadFile(lrcPath); err == nil {
		return string(data), nil
	}

	return "", os.ErrNotExist
}

// FetchLRCLIB queries the free LRCLIB public API for synchronized lyrics.
func FetchLRCLIB(artist, title, album string, duration float64) (*LyricsResult, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	baseURL := "https://lrclib.net/api/get"
	params := url.Values{}
	params.Set("artist_name", artist)
	params.Set("track_name", title)
	if album != "" {
		params.Set("album_name", album)
	}
	if duration > 0 {
		params.Set("duration", fmt.Sprintf("%.0f", duration))
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "nE-Music-Server/1.0 (https://github.com/Neutron-0/nE)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lrclibResp struct {
		SyncedLyrics string `json:"syncedLyrics"`
		PlainLyrics  string `json:"plainLyrics"`
	}

	if err := json.Unmarshal(body, &lrclibResp); err != nil {
		return nil, err
	}

	result := &LyricsResult{
		Source: "lrclib",
	}

	if lrclibResp.SyncedLyrics != "" {
		result.IsSynced = true
		result.Lines = ParseLRC(lrclibResp.SyncedLyrics)
	} else if lrclibResp.PlainLyrics != "" {
		result.PlainLyrics = lrclibResp.PlainLyrics
	}

	return result, nil
}
