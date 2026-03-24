// Package domain contains interfaces, domain types, and sentinel errors.
package domain

import "errors"

// Sentinel errors used across the service and repository layers.
// Handlers map these to HTTP status codes via mapErrorToResponse.
var (
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
	ErrDuplicateRating = errors.New("already rated this table")
	ErrVotingNotActive = errors.New("voting period is not active")
	ErrInvalidCode     = errors.New("invalid nickname or code")
	ErrFileTooLarge    = errors.New("file exceeds 10MB limit")
	ErrInvalidFileType = errors.New("only image files are allowed (jpeg, png, webp)")
)
