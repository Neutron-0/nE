package media_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ne/internal/media"
)

func TestSupportedAudioFiles(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"song.mp3", true},
		{"track.FLAC", true},
		{"audio.m4a", true},
		{"music.ogg", true},
		{"file.opus", true},
		{"sample.wav", true},
		{"audio.aiff", true},
		{"image.jpg", false},
		{"video.mp4", false},
		{"text.txt", false},
		{".neignore", false},
	}

	for _, c := range cases {
		got := media.IsSupportedAudioFile(c.path)
		if got != c.expected {
			t.Errorf("IsSupportedAudioFile(%q) = %v, expected %v", c.path, got, c.expected)
		}
	}
}

func TestArtworkDiscovery_FolderImages(t *testing.T) {
	dir := t.TempDir()
	coverPath := filepath.Join(dir, "cover.jpg")
	_ = os.WriteFile(coverPath, []byte("JPEG_IMAGE_DATA"), 0644)

	found := media.FindFolderArtwork(dir)
	if found != coverPath {
		t.Errorf("expected found artwork to be %q, got %q", coverPath, found)
	}

	// Secondary preference (folder.png)
	dir2 := t.TempDir()
	folderPng := filepath.Join(dir2, "folder.png")
	_ = os.WriteFile(folderPng, []byte("PNG_IMAGE_DATA"), 0644)

	found2 := media.FindFolderArtwork(dir2)
	if found2 != folderPng {
		t.Errorf("expected found artwork to be %q, got %q", folderPng, found2)
	}
}

func TestTranscoder_ConcurrencyAndOptions(t *testing.T) {
	tc := media.NewTranscoder(2)
	if tc.MaxConcurrent() != 2 {
		t.Errorf("expected max concurrent 2, got %d", tc.MaxConcurrent())
	}

	opts := media.TranscodeOptions{
		SourcePath:    "song.flac",
		Format:        "mp3",
		BitrateKbps:   320,
		OffsetSeconds: 15.5,
	}

	args := tc.BuildFFmpegArgs(opts)
	if len(args) == 0 {
		t.Fatal("expected non-empty args from BuildFFmpegArgs")
	}

	// Verify options format
	hasOffset := false
	hasBitrate := false
	for i, arg := range args {
		if arg == "-ss" && i+1 < len(args) && args[i+1] == "15.50" {
			hasOffset = true
		}
		if arg == "-b:a" && i+1 < len(args) && args[i+1] == "320k" {
			hasBitrate = true
		}
	}

	if !hasOffset {
		t.Errorf("expected -ss 15.50 in args, got: %v", args)
	}
	if !hasBitrate {
		t.Errorf("expected -b:a 320k in args, got: %v", args)
	}
}

func TestTranscoder_GracefulCancellation(t *testing.T) {
	tc := media.NewTranscoder(2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := tc.Stream(ctx, os.Stdout, media.TranscodeOptions{
		SourcePath:  "nonexistent.flac",
		Format:      "mp3",
		BitrateKbps: 128,
	})

	if err == nil {
		t.Error("expected error on cancelled context stream, got nil")
	}
}
