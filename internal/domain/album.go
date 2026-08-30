package domain

import "time"

type Album struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	SortTitle     string    `json:"sortTitle"`
	AlbumArtistID string    `json:"albumArtistId"`
	AlbumArtist   string    `json:"albumArtist"`
	Year          int       `json:"year"`
	OriginalYear  int       `json:"originalYear,omitempty"`
	ReleaseDate   string    `json:"releaseDate,omitempty"`
	DiscCount     int       `json:"discCount"`
	TrackCount    int       `json:"trackCount"`
	Duration      float64   `json:"duration"` // total seconds
	SizeBytes     int64     `json:"sizeBytes"`
	IsCompilation bool      `json:"isCompilation"`
	MbzAlbumID    string    `json:"mbzAlbumId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	// Hydrated fields
	Artist *Artist   `json:"artist,omitempty"`
	Tracks []*Track  `json:"tracks,omitempty"`
	Genres []string  `json:"genres,omitempty"`
	IsStarred bool   `json:"isStarred,omitempty"`
	Rating    int    `json:"rating,omitempty"`
}
