package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// RatingRepository implements domain.RatingRepository using GORM.
type RatingRepository struct {
	db *gorm.DB
}

// NewRatingRepository creates a new RatingRepository backed by the given DB.
func NewRatingRepository(db *gorm.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

// Create persists a new rating. Returns domain.ErrDuplicateRating on unique violation.
func (r *RatingRepository) Create(ctx context.Context, rating *domain.Rating) error {
	err := r.db.WithContext(ctx).Create(rating).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrDuplicateKey {
			return domain.ErrDuplicateRating
		}
		return fmt.Errorf("Create rating: %w", err)
	}
	return nil
}

// ExistsByTableAndRater returns true if the rater already rated the given table.
func (r *RatingRepository) ExistsByTableAndRater(ctx context.Context, tableID, raterID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Rating{}).
		Where("table_id = ? AND rater_id = ?", tableID, raterID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("ExistsByTableAndRater: %w", err)
	}
	return count > 0, nil
}

// GetRatedTableNumbers returns the table numbers (within a tournament) that the given rater has already rated.
func (r *RatingRepository) GetRatedTableNumbers(ctx context.Context, tournamentID, raterID uuid.UUID) ([]int, error) {
	var numbers []int
	err := r.db.WithContext(ctx).Raw(`
		SELECT t.number
		FROM ratings rt
		JOIN tables t ON rt.table_id = t.id
		WHERE t.tournament_id = ? AND rt.rater_id = ?
		ORDER BY t.number ASC
	`, tournamentID, raterID).Scan(&numbers).Error
	if err != nil {
		return nil, fmt.Errorf("GetRatedTableNumbers: %w", err)
	}
	return numbers, nil
}

// tableRow holds the raw SQL result from the aggregation query.
type tableRow struct {
	TableID     uuid.UUID `gorm:"column:table_id"`
	TableNumber int       `gorm:"column:table_number"`
	TableName   *string   `gorm:"column:table_name"`
	RatingCount int64     `gorm:"column:rating_count"`
	ZoneNone    int64     `gorm:"column:zone_none"`
	ZoneA       int64     `gorm:"column:zone_a"`
	ZoneB       int64     `gorm:"column:zone_b"`
}

// GetResultsForTournament returns aggregate results — one entry per table.
// CriteriaAverages are computed in Go from all fetched ratings.
func (r *RatingRepository) GetResultsForTournament(ctx context.Context, tournamentID uuid.UUID) ([]domain.RatingResult, error) {
	// 1. Aggregate counts per table.
	var rows []tableRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			t.id          AS table_id,
			t.number      AS table_number,
			t.name        AS table_name,
			COUNT(r.id)   AS rating_count,
			COUNT(CASE WHEN r.played_zone = 'none'   THEN 1 END) AS zone_none,
			COUNT(CASE WHEN r.played_zone = 'zone_a' THEN 1 END) AS zone_a,
			COUNT(CASE WHEN r.played_zone = 'zone_b' THEN 1 END) AS zone_b
		FROM tables t
		LEFT JOIN ratings r ON r.table_id = t.id
		WHERE t.tournament_id = ?
		GROUP BY t.id, t.number, t.name
		ORDER BY t.number ASC
	`, tournamentID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("GetResultsForTournament aggregate: %w", err)
	}

	// 2. Fetch all ratings for the tournament to compute criteria averages in Go.
	var ratings []domain.Rating
	err = r.db.WithContext(ctx).Raw(`
		SELECT r.id, r.table_id, r.rater_id, r.criteria_scores, r.played_zone, r.comment, r.created_at
		FROM ratings r
		JOIN tables t ON r.table_id = t.id
		WHERE t.tournament_id = ?
	`, tournamentID).Scan(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("GetResultsForTournament fetch ratings: %w", err)
	}

	// 3. Compute criteria averages per table in memory.
	type criteriaSum struct {
		sum   float64
		count int64
	}
	tableScores := make(map[uuid.UUID]map[string]*criteriaSum)
	for _, rating := range ratings {
		var scores map[string]int
		if err := json.Unmarshal([]byte(rating.CriteriaScores), &scores); err != nil {
			continue // skip malformed JSON
		}
		if tableScores[rating.TableID] == nil {
			tableScores[rating.TableID] = make(map[string]*criteriaSum)
		}
		for k, v := range scores {
			if tableScores[rating.TableID][k] == nil {
				tableScores[rating.TableID][k] = &criteriaSum{}
			}
			tableScores[rating.TableID][k].sum += float64(v)
			tableScores[rating.TableID][k].count++
		}
	}

	// 4. Build results.
	results := make([]domain.RatingResult, 0, len(rows))
	for _, row := range rows {
		avgs := make(map[string]float64)
		if cs, ok := tableScores[row.TableID]; ok {
			for k, s := range cs {
				if s.count > 0 {
					avgs[k] = s.sum / float64(s.count)
				}
			}
		}
		results = append(results, domain.RatingResult{
			TableID:          row.TableID,
			TableNumber:      row.TableNumber,
			TableName:        row.TableName,
			RatingCount:      row.RatingCount,
			CriteriaAverages: avgs,
			ZoneNone:         row.ZoneNone,
			ZoneA:            row.ZoneA,
			ZoneB:            row.ZoneB,
		})
	}

	// 5. Sort ascending by overall average (1 = best school grade).
	// Tables without any ratings (no "overall" key) sort last.
	sort.Slice(results, func(i, j int) bool {
		oi, hasI := results[i].CriteriaAverages["overall"]
		oj, hasJ := results[j].CriteriaAverages["overall"]
		if !hasI && !hasJ {
			return results[i].TableNumber < results[j].TableNumber
		}
		if !hasI {
			return false
		}
		if !hasJ {
			return true
		}
		return oi < oj
	})

	return results, nil
}

// GetCommentsForTournament returns all non-empty comments per table for a tournament.
// Returns map[tableID][]CommentResult — fully anonymized, no rater identity.
func (r *RatingRepository) GetCommentsForTournament(ctx context.Context, tournamentID uuid.UUID) (map[uuid.UUID][]domain.CommentResult, error) {
	type commentRow struct {
		ID       uuid.UUID `gorm:"column:id"`
		TableID  uuid.UUID `gorm:"column:table_id"`
		Comment  string    `gorm:"column:comment"`
		Approved bool      `gorm:"column:approved"`
	}
	var rows []commentRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT r.id, r.table_id, r.comment, r.approved
		FROM ratings r
		JOIN tables t ON r.table_id = t.id
		WHERE t.tournament_id = ? AND r.comment IS NOT NULL AND r.comment != ''
	`, tournamentID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("GetCommentsForTournament: %w", err)
	}
	result := make(map[uuid.UUID][]domain.CommentResult)
	for _, row := range rows {
		result[row.TableID] = append(result[row.TableID], domain.CommentResult{
			ID:       row.ID,
			Comment:  row.Comment,
			Approved: row.Approved,
		})
	}
	return result, nil
}

// SetCommentApproved updates the approved flag on a rating's comment.
func (r *RatingRepository) SetCommentApproved(ctx context.Context, ratingID uuid.UUID, approved bool) error {
	res := r.db.WithContext(ctx).Exec(`UPDATE ratings SET approved = ? WHERE id = ?`, approved, ratingID)
	if res.Error != nil {
		return fmt.Errorf("SetCommentApproved: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
