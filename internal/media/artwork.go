package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

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
