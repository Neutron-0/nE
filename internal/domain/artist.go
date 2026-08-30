package domain

import "time"

type ArtistRole string

const (
	RolePrimary  ArtistRole = "primary"
	RoleFeatured ArtistRole = "featured"
	RoleRemixer  ArtistRole = "remixer"
	RoleComposer ArtistRole = "composer"
)

type Artist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SortName    string    `json:"sortName"`
	MbzArtistID string    `json:"mbzArtistId,omitempty"`
	Biography   string    `json:"biography,omitempty"`
	AlbumCount  int       `json:"albumCount"`
	TrackCount  int       `json:"trackCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Optional hydrated user state
	IsStarred bool `json:"isStarred,omitempty"`
	Rating    int  `json:"rating,omitempty"`
}

type TrackArtist struct {
	TrackID  string     `json:"trackId"`
	ArtistID string     `json:"artistId"`
	Role     ArtistRole `json:"role"`
	Position int        `json:"position"`
	Artist   *Artist    `json:"artist,omitempty"`
}

type ParsedArtist struct {
	Name string     `json:"name"`
	Role ArtistRole `json:"role"`
}
