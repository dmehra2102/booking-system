package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeConfict      ErrorType = "CONFLICT"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrorTypeExternal     ErrorType = "EXTERNAL_ERROR"
)

type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Code    int       `json:"code"`
	Err     error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    http.StatusBadRequest,
		Err:     err,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    http.StatusNotFound,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConfict,
		Message: message,
		Code:    http.StatusConflict,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Code:    http.StatusForbidden,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    http.StatusInternalServerError,
		Err:     err,
	}
}

func NewExternalError(service, message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
		Code:    http.StatusBadGateway,
		Err:     err,
	}
}

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrDuplicateResource  = errors.New("resource already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
)

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return NewInternalError("Unknown error occurred", err)
}
