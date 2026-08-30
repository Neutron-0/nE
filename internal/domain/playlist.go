package domain

import "time"

type Playlist struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Comment    string           `json:"comment,omitempty"`
	OwnerID    string           `json:"ownerId"`
	OwnerName  string           `json:"ownerName,omitempty"`
	IsPublic   bool             `json:"isPublic"`
	Duration   float64          `json:"duration"`
	TrackCount int              `json:"trackCount"`
	Tracks     []*PlaylistTrack `json:"tracks,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`

	// Hydrated user state
	IsStarred bool `json:"isStarred,omitempty"`
}

type PlaylistTrack struct {
	ID         string    `json:"id"`
	PlaylistID string    `json:"playlistId"`
	TrackID    string    `json:"trackId"`
	Position   int       `json:"position"`
	AddedAt    time.Time `json:"addedAt"`
	Track      *Track    `json:"track,omitempty"`
}
