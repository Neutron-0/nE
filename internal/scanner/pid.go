package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"ne/internal/media"
	"github.com/google/uuid"
)

// ComputeMetadataHash creates a deterministic hash of audio identity attributes.
func ComputeMetadataHash(artist, album, title string, disc, track int, duration float64) string {
	roundedDuration := math.Round(duration)
	input := fmt.Sprintf("%s|%s|%d|%d|%s|%.0f",
		NormalizeString(strings.ToLower(artist)),
		NormalizeString(strings.ToLower(album)),
		disc,
		track,
		NormalizeString(strings.ToLower(title)),
		roundedDuration,
	)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// ComputePathHash creates a SHA-256 hash of the normalized file path.
func ComputePathHash(path string) string {
	hash := sha256.Sum256([]byte(filepathClean(path)))
	return hex.EncodeToString(hash[:])
}

func filepathClean(path string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(path)), "\\", "/")
}

// GenerateNewPID creates a fresh UUID string for a newly discovered track.
func GenerateNewPID() string {
	return uuid.NewString()
}

// IdentityResolutionResult holds resolution details for a scanned file.
type IdentityResolutionResult struct {
	PID        string
	Tier       int // 1..5
	IsExisting bool
}

// ResolveIdentity evaluates a track against existing database records using the 5-tier strategy.
func ResolveIdentity(
	path string,
	mtime int64,
	meta *media.AudioMetadata,
	lookupExistingPath func(path string) (trackID string, pid string, dbMtime int64, err error),
	lookupByMBID func(mbid string) (trackID string, pid string, err error),
	lookupByMetadataHash func(hash string, duration float64) (trackID string, pid string, err error),
) IdentityResolutionResult {
	// Tier 1 & 2: Exact or updated path match
	if trackID, pid, dbMtime, err := lookupExistingPath(path); err == nil && trackID != "" {
		if dbMtime == mtime {
			return IdentityResolutionResult{PID: pid, Tier: 1, IsExisting: true}
		}
		return IdentityResolutionResult{PID: pid, Tier: 2, IsExisting: true}
	}

	// Tier 3: MusicBrainz Track ID match
	if meta.MbzTrackID != "" {
		if trackID, pid, err := lookupByMBID(meta.MbzTrackID); err == nil && trackID != "" {
			return IdentityResolutionResult{PID: pid, Tier: 3, IsExisting: true}
		}
	}

	// Tier 4: Metadata Tuple match (duration delta < 2.0s)
	metaHash := ComputeMetadataHash(meta.Artist, meta.Album, meta.Title, meta.DiscNumber, meta.TrackNumber, meta.Duration)
	if trackID, pid, err := lookupByMetadataHash(metaHash, meta.Duration); err == nil && trackID != "" {
		return IdentityResolutionResult{PID: pid, Tier: 4, IsExisting: true}
	}

	// Tier 5: New Track Record
	return IdentityResolutionResult{
		PID:        GenerateNewPID(),
		Tier:       5,
		IsExisting: false,
	}
}
