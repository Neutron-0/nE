package domain

import "time"

type ScanStatus string

const (
	ScanStatusIdle      ScanStatus = "idle"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
	ScanStatusCancelled ScanStatus = "cancelled"
)

type ScanRun struct {
	ID             string     `json:"id"`
	LibraryID      string     `json:"libraryId"`
	Status         ScanStatus `json:"status"`
	StartedAt      time.Time  `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	FilesFound     int        `json:"filesFound"`
	FilesProcessed int        `json:"filesProcessed"`
	FilesFailed    int        `json:"filesFailed"`
	ErrorMessage   string     `json:"errorMessage,omitempty"`
}
