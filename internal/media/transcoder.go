package media

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type TranscodeOptions struct {
	SourcePath    string
	Format        string  // "mp3", "opus", "aac"
	BitrateKbps   int     // e.g. 128, 192, 320
	OffsetSeconds float64 // e.g. 30.5 for seeking
}

type Transcoder struct {
	ffmpegPath    string
	maxConcurrent int
	semaphore     chan struct{}
	mu            sync.Mutex
}

func NewTranscoder(maxConcurrent int) *Transcoder {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	ffmpegPath, _ := exec.LookPath("ffmpeg")
	return &Transcoder{
		ffmpegPath:    ffmpegPath,
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

func (t *Transcoder) IsAvailable() bool {
	return t.ffmpegPath != ""
}

func (t *Transcoder) MaxConcurrent() int {
	return t.maxConcurrent
}

func (t *Transcoder) BuildFFmpegArgs(opts TranscodeOptions) []string {
	targetCodec := "libmp3lame"
	targetContainer := "mp3"
	if opts.Format == "opus" {
		targetCodec = "libopus"
		targetContainer = "opus"
	} else if opts.Format == "aac" {
		targetCodec = "aac"
		targetContainer = "adts"
	}

	bitrate := opts.BitrateKbps
	if bitrate <= 0 {
		bitrate = 192
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
	}

	if opts.OffsetSeconds > 0 {
		args = append(args, "-ss", strconv.FormatFloat(opts.OffsetSeconds, 'f', 2, 64))
	}

	args = append(args,
		"-i", opts.SourcePath,
		"-vn",
		"-c:a", targetCodec,
		"-b:a", fmt.Sprintf("%dk", bitrate),
		"-f", targetContainer,
		"pipe:1",
	)

	return args
}

func (t *Transcoder) Stream(ctx context.Context, w io.Writer, opts TranscodeOptions) error {
	if !t.IsAvailable() {
		return fmt.Errorf("ffmpeg is not installed or available on PATH")
	}

	// Acquire concurrency token
	select {
	case t.semaphore <- struct{}{}:
		defer func() { <-t.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	args := t.BuildFFmpegArgs(opts)
	cmd := exec.CommandContext(ctx, t.ffmpegPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	_, copyErr := io.Copy(w, stdout)
	_ = cmd.Wait()

	if copyErr != nil && !strings.Contains(copyErr.Error(), "broken pipe") {
		return copyErr
	}
	return nil
}
