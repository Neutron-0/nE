package domain

import "time"

type Library struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	LastScannedAt *time.Time `json:"lastScannedAt,omitempty"`
	ScanStatus   string     `json:"scanStatus"` // "idle", "scanning", "failed"
	TrackCount   int        `json:"trackCount"`
	AlbumCount   int        `json:"albumCount"`
	TotalBytes   int64      `json:"totalBytes"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
