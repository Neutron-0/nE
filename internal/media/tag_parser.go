package media

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

type AudioMetadata struct {
	Title            string
	Artist           string
	Album            string
	AlbumArtist      string
	Composer         string
	Genre            string
	Year             int
	TrackNumber      int
	TotalTracks      int
	DiscNumber       int
	TotalDiscs       int
	Duration         float64 // in seconds
	BitRate          int     // in kbps
	SampleRate       int     // in Hz
	BitDepth         *int
	Channels         int
	Format           string
	Codec            string
	FileSize         int64
	HasEmbeddedCover bool
	MbzTrackID       string
	MbzAlbumID       string
	MbzArtistID      string
	RGTrackGain      *float64
	RGTrackPeak      *float64
	RGAlbumGain      *float64
	RGAlbumPeak      *float64
}

var (
	trackNumRegex = regexp.MustCompile(`^(\d{1,3})[\s\.\-_]+(.*)$`)
	extSupported  = map[string]string{
		".mp3":  "mp3",
		".flac": "flac",
		".m4a":  "m4a",
		".aac":  "aac",
		".ogg":  "ogg",
		".oga":  "ogg",
		".opus": "opus",
		".wav":  "wav",
		".aiff": "aiff",
		".aif":  "aiff",
		".wma":  "wma",
	}
)

// IsSupportedAudioFile checks if the given path has a supported audio extension.
func IsSupportedAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := extSupported[ext]
	return ok
}

// ParseAudioFile extracts audio metadata from an audio file.
func ParseAudioFile(path string) (*AudioMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening audio file %q: %w", path, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stating audio file %q: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	formatName := extSupported[ext]
	if formatName == "" {
		formatName = strings.TrimPrefix(ext, ".")
	}

	meta := &AudioMetadata{
		Format:      formatName,
		Codec:       formatName,
		FileSize:    stat.Size(),
		TrackNumber: 1,
		DiscNumber:  1,
		Channels:    2,
	}

	// Try extracting tags with dhowden/tag
	m, err := tag.ReadFrom(file)
	if err == nil && m != nil {
		meta.Title = strings.TrimSpace(m.Title())
		meta.Artist = strings.TrimSpace(m.Artist())
		meta.Album = strings.TrimSpace(m.Album())
		meta.AlbumArtist = strings.TrimSpace(m.AlbumArtist())
		meta.Composer = strings.TrimSpace(m.Composer())
		meta.Genre = strings.TrimSpace(m.Genre())
		meta.Year = m.Year()

		trackNo, totalTracks := m.Track()
		if trackNo > 0 {
			meta.TrackNumber = trackNo
		}
		if totalTracks > 0 {
			meta.TotalTracks = totalTracks
		}

		discNo, totalDiscs := m.Disc()
		if discNo > 0 {
			meta.DiscNumber = discNo
		}
		if totalDiscs > 0 {
			meta.TotalDiscs = totalDiscs
		}

		meta.HasEmbeddedCover = m.Picture() != nil

		if f := m.FileType(); f != "" {
			meta.Format = strings.ToLower(string(f))
			meta.Codec = strings.ToLower(string(f))
		}

		// Extract raw tags for MusicBrainz and ReplayGain if available
		raw := m.Raw()
		for k, v := range raw {
			upperK := strings.ToUpper(k)
			valStr := fmt.Sprint(v)

			if strings.Contains(upperK, "MUSICBRAINZ_TRACKID") || strings.Contains(upperK, "UFID") {
				meta.MbzTrackID = valStr
			}
			if strings.Contains(upperK, "MUSICBRAINZ_ALBUMID") {
				meta.MbzAlbumID = valStr
			}
			if strings.Contains(upperK, "MUSICBRAINZ_ARTISTID") {
				meta.MbzArtistID = valStr
			}
			if strings.Contains(upperK, "REPLAYGAIN_TRACK_GAIN") {
				if g, err := strconv.ParseFloat(strings.TrimSuffix(valStr, " dB"), 64); err == nil {
					meta.RGTrackGain = &g
				}
			}
			if strings.Contains(upperK, "REPLAYGAIN_TRACK_PEAK") {
				if p, err := strconv.ParseFloat(valStr, 64); err == nil {
					meta.RGTrackPeak = &p
				}
			}
			if strings.Contains(upperK, "REPLAYGAIN_ALBUM_GAIN") {
				if g, err := strconv.ParseFloat(strings.TrimSuffix(valStr, " dB"), 64); err == nil {
					meta.RGAlbumGain = &g
				}
			}
			if strings.Contains(upperK, "REPLAYGAIN_ALBUM_PEAK") {
				if p, err := strconv.ParseFloat(valStr, 64); err == nil {
					meta.RGAlbumPeak = &p
				}
			}
		}
	}

	// Heuristic fallbacks for missing title / artist / album
	applyFallbacks(meta, path)

	// Approximate duration if unavailable: estimation based on filesize and standard bitrate
	if meta.Duration <= 0 {
		meta.Duration = estimateDuration(meta.FileSize, meta.Format)
	}

	return meta, nil
}

func applyFallbacks(meta *AudioMetadata, path string) {
	filename := filepath.Base(path)
	filenameNoExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	cleanName := filenameNoExt

	// 1. Pattern matching for Title and Artist from filename (e.g. "Moonlight - Kali Uchis")
	if strings.Contains(cleanName, " - ") {
		parts := strings.Split(cleanName, " - ")
		if len(parts) == 2 {
			p0 := strings.TrimSpace(parts[0])
			p1 := strings.TrimSpace(parts[1])

			if matches := trackNumRegex.FindStringSubmatch(p0); len(matches) == 3 {
				if num, err := strconv.Atoi(matches[1]); err == nil && meta.TrackNumber <= 1 {
					meta.TrackNumber = num
				}
				p0 = strings.TrimSpace(matches[2])
			}

			if meta.Title == "" || meta.Title == filenameNoExt {
				meta.Title = p0
			}
			if meta.Artist == "" || meta.Artist == "music" {
				meta.Artist = p1
			}
		} else if len(parts) >= 3 {
			if num, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil && meta.TrackNumber <= 1 {
				meta.TrackNumber = num
			}
			if meta.Artist == "" || meta.Artist == "music" {
				meta.Artist = strings.TrimSpace(parts[1])
			}
			if meta.Title == "" || meta.Title == filenameNoExt {
				meta.Title = strings.TrimSpace(parts[2])
			}
		}
	} else if meta.Title == "" {
		matches := trackNumRegex.FindStringSubmatch(cleanName)
		if len(matches) == 3 {
			if num, err := strconv.Atoi(matches[1]); err == nil && meta.TrackNumber <= 1 {
				meta.TrackNumber = num
			}
			meta.Title = strings.TrimSpace(matches[2])
		} else {
			meta.Title = cleanName
		}
	}

	// 2. Filter out generic root folder names
	dir := filepath.Dir(path)
	parentDir := filepath.Base(dir)
	grandparentDir := filepath.Base(filepath.Dir(dir))

	isRootMusicDir := func(d string) bool {
		lower := strings.ToLower(strings.TrimSpace(d))
		return lower == "music" || lower == "musicdir" || lower == "." || lower == "/" || lower == "\\" || lower == "app" || lower == "songs" || lower == "mp3"
	}

	if meta.Artist == "" || isRootMusicDir(meta.Artist) {
		if !isRootMusicDir(parentDir) {
			meta.Artist = parentDir
		} else if !isRootMusicDir(grandparentDir) {
			meta.Artist = grandparentDir
		} else {
			meta.Artist = "Unknown Artist"
		}
	}

	// 3. Album fallback
	if meta.Album == "" || isRootMusicDir(meta.Album) {
		if !isRootMusicDir(parentDir) {
			meta.Album = parentDir
			if !isRootMusicDir(grandparentDir) && (meta.Artist == "" || isRootMusicDir(meta.Artist)) {
				meta.Artist = grandparentDir
			}
		} else {
			meta.Album = "Singles"
		}
	}

	if meta.AlbumArtist == "" || isRootMusicDir(meta.AlbumArtist) {
		meta.AlbumArtist = meta.Artist
	}
}

func estimateDuration(fileSize int64, format string) float64 {
	bitrateKbps := 320.0
	switch strings.ToLower(format) {
	case "flac", "wav", "aiff":
		bitrateKbps = 800.0
	case "mp3", "m4a", "aac":
		bitrateKbps = 256.0
	case "ogg", "opus":
		bitrateKbps = 160.0
	}
	bytesPerSecond := (bitrateKbps * 1000.0) / 8.0
	return float64(fileSize) / bytesPerSecond
}
