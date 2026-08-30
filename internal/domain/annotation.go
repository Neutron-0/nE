package domain

import "time"

type UserTrackAnnotation struct {
	UserID       string     `json:"userId"`
	TrackID      string     `json:"trackId"`
	IsStarred    bool       `json:"isStarred"`
	StarredAt    *time.Time `json:"starredAt,omitempty"`
	Rating       int        `json:"rating"` // 0-5
	RatedAt      *time.Time `json:"ratedAt,omitempty"`
	PlayCount    int        `json:"playCount"`
	LastPlayedAt *time.Time `json:"lastPlayedAt,omitempty"`
}

type UserAlbumAnnotation struct {
	UserID       string     `json:"userId"`
	AlbumID      string     `json:"albumId"`
	IsStarred    bool       `json:"isStarred"`
	StarredAt    *time.Time `json:"starredAt,omitempty"`
	Rating       int        `json:"rating"` // 0-5
	RatedAt      *time.Time `json:"ratedAt,omitempty"`
	PlayCount    int        `json:"playCount"`
	LastPlayedAt *time.Time `json:"lastPlayedAt,omitempty"`
}

type UserArtistAnnotation struct {
	UserID    string     `json:"userId"`
	ArtistID  string     `json:"artistId"`
	IsStarred bool       `json:"isStarred"`
	StarredAt *time.Time `json:"starredAt,omitempty"`
	Rating    int        `json:"rating"` // 0-5
	RatedAt   *time.Time `json:"ratedAt,omitempty"`
}
