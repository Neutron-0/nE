package domain

import "time"

// User represents an authenticated account in nE.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	IsAdmin      bool      `json:"isAdmin"`
	CanTranscode bool      `json:"canTranscode"`
	TokenVersion int       `json:"-"`
	SubsonicSalt string    `json:"-"`
	SubsonicToken string   `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// RefreshToken stores a hashed session token for JWT rotation.
type RefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}

func (r *RefreshToken) IsActive() bool {
	return r.RevokedAt == nil && time.Now().Before(r.ExpiresAt)
}
