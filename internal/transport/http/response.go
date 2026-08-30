package http

import (
	"encoding/json"
	"net/http"

	"ne/internal/domain"
	"github.com/go-chi/chi/v5/middleware"
)

// ErrorResponse represents the standardized JSON error envelope.
type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code      domain.ErrorCode  `json:"code"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

// CollectionResponse represents the standardized collection envelope.
type CollectionResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
	Total      int    `json:"total,omitempty"`
}

// JSON sends a JSON response with status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// RespondError maps an error to the standard JSON error envelope and appropriate HTTP status.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetReqID(r.Context())
	
	appErr, ok := err.(*domain.AppError)
	if !ok {
		appErr = domain.ErrInternal(err)
	}

	payload := ErrorResponse{
		Error: ErrorPayload{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Details:   appErr.Details,
			RequestID: reqID,
		},
	}

	status := appErr.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}

	JSON(w, status, payload)
}
