package domain

import "time"

type Track struct {
	ID               string    `json:"id"`
	PID              string    `json:"pid"`
	LibraryID        string    `json:"libraryId"`
	Path             string    `json:"path"`
	FolderPath       string    `json:"folderPath"`
	Filename         string    `json:"filename"`
	Title            string    `json:"title"`
	SortTitle        string    `json:"sortTitle"`
	RawArtist        string    `json:"rawArtist"`
	AlbumID          string    `json:"albumId"`
	TrackNumber      int       `json:"trackNumber"`
	DiscNumber       int       `json:"discNumber"`
	DiscSubtitle     string    `json:"discSubtitle,omitempty"`
	Year             int       `json:"year"`
	Duration         float64   `json:"duration"` // seconds
	BitRate          int       `json:"bitRate"`  // kbps
	SampleRate       int       `json:"sampleRate"`
	BitDepth         *int      `json:"bitDepth,omitempty"`
	Channels         int       `json:"channels"`
	Format           string    `json:"format"`
	Codec            string    `json:"codec"`
	FileSize         int64     `json:"fileSize"`
	RGTrackGain      *float64  `json:"rgTrackGain,omitempty"`
	RGTrackPeak      *float64  `json:"rgTrackPeak,omitempty"`
	RGAlbumGain      *float64  `json:"rgAlbumGain,omitempty"`
	RGAlbumPeak      *float64  `json:"rgAlbumPeak,omitempty"`
	HasEmbeddedCover bool      `json:"hasEmbeddedCover"`
	MbzTrackID       string    `json:"mbzTrackId,omitempty"`
	MTime            int64     `json:"mtime"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`

	// Hydrated fields
	AlbumArtist string         `json:"albumArtist,omitempty"`
	AlbumTitle  string         `json:"albumTitle,omitempty"`
	Artists     []*TrackArtist `json:"artists,omitempty"`
	Genres      []string       `json:"genres,omitempty"`
	IsStarred   bool           `json:"isStarred,omitempty"`
	Rating      int            `json:"rating,omitempty"`
	PlayCount   int            `json:"playCount,omitempty"`
}
