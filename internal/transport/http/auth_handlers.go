package http

import (
	"encoding/json"
	"net/http"
	"time"

	"ne/internal/domain"
	"ne/internal/service"
	"github.com/go-chi/chi/v5"
)

const RefreshCookieName = "ne_refresh_token"

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/setup", h.Setup)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)

	r.Group(func(protected chi.Router) {
		protected.Use(RequireAuth)
		protected.Get("/me", h.Me)
		protected.Post("/change-password", h.ChangePassword)
	})

	return r
}

type SetupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	result, err := h.authService.SetupInitialAdmin(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	setRefreshCookie(w, result.RefreshToken, 30*24*time.Hour)
	JSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	result, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	setRefreshCookie(w, result.RefreshToken, 30*24*time.Hour)
	JSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookie, err := r.Cookie(RefreshCookieName); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		var req RefreshRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		RespondError(w, r, domain.ErrUnauthorized("Missing refresh token"))
		return
	}

	result, err := h.authService.Refresh(r.Context(), refreshToken)
	if err != nil {
		clearRefreshCookie(w)
		RespondError(w, r, err)
		return
	}

	setRefreshCookie(w, result.RefreshToken, 30*24*time.Hour)
	JSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookie, err := r.Cookie(RefreshCookieName); err == nil {
		refreshToken = cookie.Value
	}

	_ = h.authService.Logout(r.Context(), refreshToken)
	clearRefreshCookie(w)
	JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())
	user, err := h.authService.GetCurrentUser(r.Context(), claims.UserID)
	if err != nil {
		RespondError(w, r, err)
		return
	}
	JSON(w, http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, domain.ErrInvalidInput("Invalid JSON request body"))
		return
	}

	if err := h.authService.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		RespondError(w, r, err)
		return
	}

	clearRefreshCookie(w)
	JSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
}

func setRefreshCookie(w http.ResponseWriter, token string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   false, // set to true when served over HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(maxAge.Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
