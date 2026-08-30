package scanner_test

import (
	"errors"
	"testing"

	"ne/internal/media"
	"ne/internal/scanner"
)

func TestPID_GenerationAndFormat(t *testing.T) {
	pid1 := scanner.GenerateNewPID()
	pid2 := scanner.GenerateNewPID()

	if len(pid1) != 36 { // UUID standard format
		t.Errorf("expected 36-char UUID PID, got %d chars: %q", len(pid1), pid1)
	}
	if len(pid2) != 36 {
		t.Errorf("expected 36-char UUID PID, got %d chars: %q", len(pid2), pid2)
	}
	if pid1 == pid2 {
		t.Errorf("expected distinct PIDs, got collision: %s == %s", pid1, pid2)
	}
}

func TestPID_MetadataHash(t *testing.T) {
	h1 := scanner.ComputeMetadataHash("Radiohead", "In Rainbows", "15 Step", 1, 1, 237.0)
	h2 := scanner.ComputeMetadataHash("Radiohead", "In Rainbows", "15 Step", 1, 1, 237.4) // rounds to 237
	h3 := scanner.ComputeMetadataHash("Radiohead", "In Rainbows", "Bodysnatchers", 1, 2, 242.0)

	if h1 != h2 {
		t.Errorf("expected identical hash for close durations, got %s != %s", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("expected distinct hash for different tracks, got collision: %s", h1)
	}
}

func TestPID_5TierResolution(t *testing.T) {
	meta := &media.AudioMetadata{
		Artist:      "Radiohead",
		Album:       "In Rainbows",
		Title:       "15 Step",
		TrackNumber: 1,
		DiscNumber:  1,
		Duration:    237.0,
		MbzTrackID:  "mbz-rec-12345",
	}

	// 1. Tier 1 Test (Exact path + mtime)
	r1 := scanner.ResolveIdentity(
		"/music/01.flac",
		1000,
		meta,
		func(path string) (string, string, int64, error) {
			return "trk-1", "pid-1", 1000, nil
		},
		func(mbid string) (string, string, error) { return "", "", errors.New("not found") },
		func(hash string, dur float64) (string, string, error) { return "", "", errors.New("not found") },
	)
	if r1.Tier != 1 || r1.PID != "pid-1" || !r1.IsExisting {
		t.Errorf("expected Tier 1 exact match, got %+v", r1)
	}

	// 2. Tier 2 Test (Same path, modified mtime)
	r2 := scanner.ResolveIdentity(
		"/music/01.flac",
		1050,
		meta,
		func(path string) (string, string, int64, error) {
			return "trk-1", "pid-1", 1000, nil
		},
		func(mbid string) (string, string, error) { return "", "", errors.New("not found") },
		func(hash string, dur float64) (string, string, error) { return "", "", errors.New("not found") },
	)
	if r2.Tier != 2 || r2.PID != "pid-1" || !r2.IsExisting {
		t.Errorf("expected Tier 2 modified match, got %+v", r2)
	}

	// 3. Tier 3 Test (Different path, MBID match)
	r3 := scanner.ResolveIdentity(
		"/music/renamed/01.flac",
		1050,
		meta,
		func(path string) (string, string, int64, error) { return "", "", 0, errors.New("not found") },
		func(mbid string) (string, string, error) {
			if mbid == "mbz-rec-12345" {
				return "trk-1", "pid-1", nil
			}
			return "", "", errors.New("not found")
		},
		func(hash string, dur float64) (string, string, error) { return "", "", errors.New("not found") },
	)
	if r3.Tier != 3 || r3.PID != "pid-1" || !r3.IsExisting {
		t.Errorf("expected Tier 3 MBID match, got %+v", r3)
	}

	// 4. Tier 4 Test (Different path, no MBID, matching metadata tuple)
	metaNoMBID := *meta
	metaNoMBID.MbzTrackID = ""
	r4 := scanner.ResolveIdentity(
		"/music/renamed/01.flac",
		1050,
		&metaNoMBID,
		func(path string) (string, string, int64, error) { return "", "", 0, errors.New("not found") },
		func(mbid string) (string, string, error) { return "", "", errors.New("not found") },
		func(hash string, dur float64) (string, string, error) {
			return "trk-1", "pid-1", nil
		},
	)
	if r4.Tier != 4 || r4.PID != "pid-1" || !r4.IsExisting {
		t.Errorf("expected Tier 4 metadata match, got %+v", r4)
	}

	// 5. Tier 5 Test (Brand new track)
	r5 := scanner.ResolveIdentity(
		"/music/new/song.flac",
		2000,
		&metaNoMBID,
		func(path string) (string, string, int64, error) { return "", "", 0, errors.New("not found") },
		func(mbid string) (string, string, error) { return "", "", errors.New("not found") },
		func(hash string, dur float64) (string, string, error) { return "", "", errors.New("not found") },
	)
	if r5.Tier != 5 || r5.PID == "" || r5.IsExisting {
		t.Errorf("expected Tier 5 new track, got %+v", r5)
	}
}
