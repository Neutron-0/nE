package domain

import "time"

type PlaybackRecord struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	TrackID        string    `json:"trackId"`
	PlayerName     string    `json:"playerName,omitempty"`
	PlayedAt       time.Time `json:"playedAt"`
	DurationPlayed float64   `json:"durationPlayed"`
	Completed      bool      `json:"completed"`
	Track          *Track    `json:"track,omitempty"`
}
