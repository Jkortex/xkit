package handler

import (
	"errors"
	"net/http"

	"daily/internal/application/apperr"
)

func statusFromError(err error) int {
	switch {
	case errors.Is(err, apperr.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, apperr.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperr.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
