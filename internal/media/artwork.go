package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

type ExtractedArtwork struct {
	Data        []byte
	MimeType    string
	Fingerprint string
}

// FindFolderArtwork searches a directory for common album cover filenames.
func FindFolderArtwork(dir string) string {
	candidates := []string{
		"cover.jpg", "cover.png", "cover.jpeg", "cover.webp",
		"folder.jpg", "folder.png", "folder.jpeg",
		"front.jpg", "front.png", "front.jpeg",
		"album.jpg", "album.png",
	}

	for _, cand := range candidates {
		candPath := filepath.Join(dir, cand)
		if stat, err := os.Stat(candPath); err == nil && !stat.IsDir() && stat.Size() > 0 {
			return candPath
		}
	}
	return ""
}

// ExtractArtwork attempts to extract embedded album artwork or scans the parent directory for cover art images.
func ExtractArtwork(filePath string) (*ExtractedArtwork, error) {
	// 1. Check embedded cover art first
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()
		m, err := tag.ReadFrom(file)
		if err == nil && m.Picture() != nil {
			pic := m.Picture()
			if len(pic.Data) > 0 {
				hash := sha256.Sum256(pic.Data)
				return &ExtractedArtwork{
					Data:        pic.Data,
					MimeType:    pic.MIMEType,
					Fingerprint: hex.EncodeToString(hash[:]),
				}, nil
			}
		}
	}

	// 2. Fallback: Search directory for cover images
	folderCover := FindFolderArtwork(filepath.Dir(filePath))
	if folderCover != "" {
		if data, err := os.ReadFile(folderCover); err == nil && len(data) > 0 {
			mimeType := "image/jpeg"
			if strings.HasSuffix(strings.ToLower(folderCover), ".png") {
				mimeType = "image/png"
			} else if strings.HasSuffix(strings.ToLower(folderCover), ".webp") {
				mimeType = "image/webp"
			}
			hash := sha256.Sum256(data)
			return &ExtractedArtwork{
				Data:        data,
				MimeType:    mimeType,
				Fingerprint: hex.EncodeToString(hash[:]),
			}, nil
		}
	}

	return nil, errors.New("no artwork found")
}

// GetImageDimensions returns width and height for valid image data.
func GetImageDimensions(data []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// FetchOnlineArtwork queries public metadata services to find cover art for tracks without embedded artwork.
func FetchOnlineArtwork(artist, title string) (*ExtractedArtwork, error) {
	cleanArtist := strings.TrimSpace(artist)
	cleanTitle := strings.TrimSpace(title)
	if cleanArtist == "" && cleanTitle == "" {
		return nil, errors.New("insufficient metadata for online artwork search")
	}

	term := url.QueryEscape(fmt.Sprintf("%s %s", cleanArtist, cleanTitle))
	reqURL := fmt.Sprintf("https://itunes.apple.com/search?term=%s&entity=song&limit=1", term)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var itunesRes struct {
		ResultCount int `json:"resultCount"`
		Results     []struct {
			ArtworkUrl100 string `json:"artworkUrl100"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&itunesRes); err != nil || itunesRes.ResultCount == 0 {
		return nil, errors.New("no online artwork found")
	}

	hiresURL := strings.Replace(itunesRes.Results[0].ArtworkUrl100, "100x100bb.jpg", "600x600bb.jpg", 1)
	imgResp, err := client.Get(hiresURL)
	if err != nil {
		return nil, err
	}
	defer imgResp.Body.Close()

	data, err := io.ReadAll(imgResp.Body)
	if err != nil || len(data) == 0 {
		return nil, errors.New("failed reading online image data")
	}

	hash := sha256.Sum256(data)
	return &ExtractedArtwork{
		Data:        data,
		MimeType:    "image/jpeg",
		Fingerprint: hex.EncodeToString(hash[:]),
	}, nil
}
