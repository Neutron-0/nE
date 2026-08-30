package domain

import "fmt"

// ErrorCode defines application-wide machine-readable error codes.
type ErrorCode string

const (
	ErrCodeInternal           ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeInvalidRequest     ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeConflict           ErrorCode = "CONFLICT"
	ErrCodeStorageUnavailable ErrorCode = "STORAGE_UNAVAILABLE"
	ErrCodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
)

// AppError represents a structured domain error.
type AppError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code ErrorCode, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
		Details:    make(map[string]string),
	}
}

// Common error constructors
func ErrNotFound(resource string, id string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s with id %q was not found", resource, id),
		HTTPStatus: 404,
		Details:    map[string]string{"resource": resource, "id": id},
	}
}

func ErrInvalidInput(message string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidRequest,
		Message:    message,
		HTTPStatus: 400,
	}
}

func ErrConflict(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		HTTPStatus: 409,
	}
}

func ErrUnauthorized(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: 401,
	}
}

func ErrForbidden(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: 403,
	}
}

func ErrInternal(err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    "An unexpected internal error occurred",
		HTTPStatus: 500,
		Err:        err,
	}
}
