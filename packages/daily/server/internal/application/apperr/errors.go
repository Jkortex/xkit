package apperr

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrInvalidTransition  = errors.New("invalid transition")
	ErrPreconditionFailed = errors.New("precondition failed")
)
