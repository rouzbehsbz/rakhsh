package common

import (
	"errors"
	"net/http"
)

var (
	ErrForbidden        = errors.New("Forbidden")
	ErrNotFound         = errors.New("Resource not found")
	ErrInternalDatabase = errors.New("Internal database error")
	ErrInternal         = errors.New("Oops ! something went wrong")
)

type AppError struct {
	StatusCode int
	Message    string
}

func (a *AppError) Error() string {
	return a.Message
}

func NotFoundError(message string) *AppError {
	if message == "" {
		message = ErrNotFound.Error()
	}

	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func InternalError(message string) *AppError {
	if message == "" {
		message = ErrInternal.Error()
	}

	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

func ForbiddenError(message string) *AppError {
	if message == "" {
		message = ErrForbidden.Error()
	}

	return &AppError{
		StatusCode: http.StatusMethodNotAllowed,
		Message:    message,
	}
}
