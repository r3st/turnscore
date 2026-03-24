package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/r3st/turnscore/internal/domain"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	errs := []error{
		domain.ErrNotFound,
		domain.ErrUnauthorized,
		domain.ErrForbidden,
		domain.ErrConflict,
		domain.ErrDuplicateRating,
		domain.ErrVotingNotActive,
		domain.ErrInvalidCode,
		domain.ErrFileTooLarge,
		domain.ErrInvalidFileType,
	}
	for i, a := range errs {
		for j, b := range errs {
			if i != j {
				assert.False(t, errors.Is(a, b), "%v should not match %v", a, b)
			}
		}
	}
}

func TestSentinelErrors_WrappedErrorsMatch(t *testing.T) {
	wrapped := fmt.Errorf("tournament service: %w", domain.ErrNotFound)
	assert.True(t, errors.Is(wrapped, domain.ErrNotFound))
	assert.False(t, errors.Is(wrapped, domain.ErrForbidden))
}
