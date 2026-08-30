package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type UserClaims struct {
	UserID       string
	Username     string
	IsAdmin      bool
	TokenVersion int
}

type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTManager(secretKey string, configDir string, accessExpiry, refreshExpiry time.Duration) (*JWTManager, error) {
	var keyBytes []byte

	if secretKey != "" {
		keyBytes = []byte(secretKey)
	} else {
		// Attempt to load from file or generate persistent key
		keyFile := filepath.Join(configDir, "jwt_secret.key")
		if data, err := os.ReadFile(keyFile); err == nil && len(data) >= 32 {
			keyBytes = data
		} else {
			// Generate new 32-byte cryptographically secure key
			keyBytes = make([]byte, 32)
			if _, err := rand.Read(keyBytes); err != nil {
				return nil, fmt.Errorf("generating secret key: %w", err)
			}
			_ = os.MkdirAll(configDir, 0700)
			if err := os.WriteFile(keyFile, keyBytes, 0600); err != nil {
				return nil, fmt.Errorf("saving jwt key file: %w", err)
			}
		}
	}

	return &JWTManager{
		secretKey:     keyBytes,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateAccessToken signs a short-lived JWT containing essential user claims.
func (m *JWTManager) GenerateAccessToken(claims UserClaims) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessExpiry)

	builder := jwt.NewBuilder()
	builder.Subject(claims.UserID)
	builder.IssuedAt(now)
	builder.Expiration(expiresAt)
	builder.Claim("username", claims.Username)
	builder.Claim("is_admin", claims.IsAdmin)
	builder.Claim("token_version", claims.TokenVersion)

	tok, err := builder.Build()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("building jwt: %w", err)
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256, m.secretKey))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing jwt: %w", err)
	}

	return string(signed), expiresAt, nil
}

// VerifyAccessToken parses and validates a signed JWT token string.
func (m *JWTManager) VerifyAccessToken(tokenStr string) (*UserClaims, error) {
	tok, err := jwt.Parse([]byte(tokenStr), jwt.WithKey(jwa.HS256, m.secretKey), jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	userID := tok.Subject()
	if userID == "" {
		return nil, errors.New("missing subject claim")
	}

	username, ok := tok.Get("username")
	if !ok {
		return nil, errors.New("missing username claim")
	}

	isAdminVal, ok := tok.Get("is_admin")
	if !ok {
		return nil, errors.New("missing is_admin claim")
	}
	isAdmin, _ := isAdminVal.(bool)

	tokenVersionVal, ok := tok.Get("token_version")
	if !ok {
		return nil, errors.New("missing token_version claim")
	}
	var tokenVersion int
	switch v := tokenVersionVal.(type) {
	case float64:
		tokenVersion = int(v)
	case int:
		tokenVersion = v
	}

	return &UserClaims{
		UserID:       userID,
		Username:     fmt.Sprint(username),
		IsAdmin:      isAdmin,
		TokenVersion: tokenVersion,
	}, nil
}

// GenerateRefreshToken creates a random 32-byte opaque token string and its SHA-256 hash.
func GenerateRefreshToken() (tokenString string, tokenHash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generating random refresh token: %w", err)
	}

	tokenString = base64.RawURLEncoding.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash = hex.EncodeToString(hash[:])
	return tokenString, tokenHash, nil
}

// HashRefreshToken calculates the SHA-256 hex string of a refresh token for database lookup.
func HashRefreshToken(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}
