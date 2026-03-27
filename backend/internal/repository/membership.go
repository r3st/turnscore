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

// ListByTournamentID returns all members with user info for a tournament, ordered by name ASC.
func (r *MembershipRepository) ListByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]domain.MemberWithUser, error) {
	type row struct {
		UserID    uuid.UUID `gorm:"column:user_id"`
		Name      string    `gorm:"column:name"`
		AvatarURL *string   `gorm:"column:avatar_url"`
		Role      string    `gorm:"column:role"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT tm.user_id, u.name, u.avatar_url, tm.role
		FROM tournament_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.tournament_id = ?
		ORDER BY u.name ASC
	`, tournamentID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("ListByTournamentID: %w", err)
	}
	result := make([]domain.MemberWithUser, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.MemberWithUser{
			UserID:    row.UserID,
			Name:      row.Name,
			AvatarURL: row.AvatarURL,
			Role:      row.Role,
		})
	}
	return result, nil
}

// ListWithTournamentsByUserID returns memberships with tournament info for a user.
func (r *MembershipRepository) ListWithTournamentsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MembershipWithTournament, error) {
	type row struct {
		TournamentID   uuid.UUID `gorm:"column:tournament_id"`
		TournamentName string    `gorm:"column:tournament_name"`
		TournamentSlug string    `gorm:"column:tournament_slug"`
		Role           string    `gorm:"column:role"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT tm.tournament_id, t.name AS tournament_name, t.slug AS tournament_slug, tm.role
		FROM tournament_members tm
		JOIN tournaments t ON tm.tournament_id = t.id
		WHERE tm.user_id = ?
		ORDER BY t.name ASC
	`, userID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("ListWithTournamentsByUserID: %w", err)
	}
	result := make([]domain.MembershipWithTournament, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.MembershipWithTournament{
			TournamentID:   row.TournamentID,
			TournamentName: row.TournamentName,
			TournamentSlug: row.TournamentSlug,
			Role:           row.Role,
		})
	}
	return result, nil
}
