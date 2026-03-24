package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

const pgErrDuplicateKey = "23505"

// MembershipRepository implements domain.TournamentMemberRepository using GORM.
type MembershipRepository struct {
	db *gorm.DB
}

// NewMembershipRepository creates a new MembershipRepository.
func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// GetRole returns the role of userID in the given tournament, or domain.ErrNotFound if not a member.
func (r *MembershipRepository) GetRole(ctx context.Context, tournamentID, userID uuid.UUID) (string, error) {
	var m domain.TournamentMembership
	err := r.db.WithContext(ctx).
		Where("tournament_id = ? AND user_id = ?", tournamentID, userID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("GetRole: %w", err)
	}
	return m.Role, nil
}

// AddMember inserts a new membership record. Returns domain.ErrConflict on duplicate.
func (r *MembershipRepository) AddMember(ctx context.Context, tournamentID, userID uuid.UUID, role string) error {
	m := domain.TournamentMembership{
		TournamentID: tournamentID,
		UserID:       userID,
		Role:         role,
	}
	err := r.db.WithContext(ctx).Create(&m).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrDuplicateKey {
			return domain.ErrConflict
		}
		return fmt.Errorf("AddMember: %w", err)
	}
	return nil
}

// RemoveMember deletes the membership of userID from the given tournament.
func (r *MembershipRepository) RemoveMember(ctx context.Context, tournamentID, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Where("tournament_id = ? AND user_id = ?", tournamentID, userID).
		Delete(&domain.TournamentMembership{}).Error
	if err != nil {
		return fmt.Errorf("RemoveMember: %w", err)
	}
	return nil
}
