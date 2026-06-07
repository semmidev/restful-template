package todos

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

// allowedSortCols is an explicit allowlist to prevent SQL injection via the
// sort parameter — never interpolate user-supplied column names directly.
var allowedSortCols = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"title":      "title",
	"status":     "status",
}

type todoRepository struct{ db *pgxpool.Pool }

func NewTodoRepository(db *pgxpool.Pool) TodoRepository { return &todoRepository{db} }

func (r *todoRepository) Create(ctx context.Context, t *Todo) error {
	sql, args, err := database.QB.Insert("todos").
		Columns("id", "user_id", "title", "description", "cover", "status", "importance", "urgency", "created_at", "updated_at").
		Values(t.ID, t.UserID, t.Title, t.Description, t.Cover, t.Status, t.Importance, t.Urgency, t.CreatedAt, t.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *todoRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Todo, error) {
	sql, args, err := database.QB.Select("id", "user_id", "title", "description", "cover", "status", "importance", "urgency", "created_at", "updated_at").
		From("todos").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := database.GetDB(ctx, r.db).QueryRow(ctx, sql, args...)
	var t Todo
	if err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Cover, &t.Status, &t.Importance, &t.Urgency, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// ListByUser returns paginated todos using COUNT(*) OVER() to get the total
// in the same query, avoiding a second round-trip for pagination metadata.
func (r *todoRepository) ListByUser(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error) {
	base := database.QB.Select().From("todos").Where(sq.Eq{"user_id": q.UserID})

	if q.Status != nil {
		base = base.Where(sq.Eq{"status": *q.Status})
	}

	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		base = base.Where(sq.Or{
			sq.ILike{"title": like},
			sq.ILike{"description": like},
		})
	}

	// Fall back to the default if the caller supplied an unknown column.
	sortBy := "created_at"
	if col, ok := allowedSortCols[q.SortBy]; ok {
		sortBy = col
	}

	sortDir := "DESC"
	if q.SortDir == "asc" {
		sortDir = "ASC"
	}

	dataQuery := base.
		Columns(
			"id", "user_id", "title", "description", "cover",
			"status", "importance", "urgency", "created_at", "updated_at",
			"COUNT(*) OVER() AS total_count",
		).
		OrderBy(sortBy + " " + sortDir).
		Limit(uint64(q.Limit)).
		Offset(uint64(q.Offset))

	dataSQL, dataArgs, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := database.GetDB(ctx, r.db).Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total int
	out := make([]*Todo, 0)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description, &t.Cover,
			&t.Status, &t.Importance, &t.Urgency, &t.CreatedAt, &t.UpdatedAt,
			&total,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

// Update persists the mutated todo entity.
// Checking RowsAffected == 0 catches concurrent deletes that happen between
// the handler's ETag fetch and this write, returning ErrNotFound instead of
// silently no-op'ing.
func (r *todoRepository) Update(ctx context.Context, t *Todo) error {
	sql, args, err := database.QB.Update("todos").
		Set("title", t.Title).
		Set("description", t.Description).
		Set("cover", t.Cover).
		Set("status", t.Status).
		Set("importance", t.Importance).
		Set("urgency", t.Urgency).
		Set("updated_at", t.UpdatedAt).
		Where(sq.Eq{"id": t.ID, "user_id": t.UserID}).
		ToSql()
	if err != nil {
		return err
	}

	res, err := database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	sql, args, err := database.QB.Delete("todos").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *todoRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	sql, args, err := database.QB.Delete("todos").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *todoRepository) GetStats(ctx context.Context, userID uuid.UUID) (*TodoStats, error) {
	// 1. Get status counts
	countSQL, countArgs, err := database.QB.Select(
		"COUNT(*)",
		"COUNT(*) FILTER (WHERE status = 'pending')",
		"COUNT(*) FILTER (WHERE status = 'in_progress')",
		"COUNT(*) FILTER (WHERE status = 'done')",
	).From("todos").Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return nil, err
	}

	var stats TodoStats
	row := database.GetDB(ctx, r.db).QueryRow(ctx, countSQL, countArgs...)
	err = row.Scan(&stats.Total, &stats.Pending, &stats.InProgress, &stats.Completed)
	if err != nil {
		return nil, err
	}

	if stats.Total > 0 {
		stats.CompletionRate = int(float64(stats.Completed) / float64(stats.Total) * 100)
	} else {
		stats.CompletionRate = 0
	}

	// 2. Daily stats for the last 7 days
	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -6)

	stats.DailyStats = make([]DailyStat, 7)
	dateMap := make(map[string]int)
	for i := 0; i < 7; i++ {
		d := startDate.AddDate(0, 0, i)
		dateStr := d.Format("2006-01-02")
		stats.DailyStats[i] = DailyStat{
			Date: dateStr,
		}
		dateMap[dateStr] = i
	}

	createdSQL, createdArgs, err := database.QB.Select(
		"DATE(created_at) AS date",
		"COUNT(*)",
	).From("todos").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.GtOrEq{"created_at": startDate}).
		GroupBy("DATE(created_at)").
		ToSql()
	if err != nil {
		return nil, err
	}

	createdRows, err := database.GetDB(ctx, r.db).Query(ctx, createdSQL, createdArgs...)
	if err != nil {
		return nil, err
	}
	defer createdRows.Close()

	for createdRows.Next() {
		var dateVal time.Time
		var count int
		if err := createdRows.Scan(&dateVal, &count); err != nil {
			return nil, err
		}
		dateStr := dateVal.Format("2006-01-02")
		if idx, ok := dateMap[dateStr]; ok {
			stats.DailyStats[idx].Created = count
		}
	}
	if err := createdRows.Err(); err != nil {
		return nil, err
	}

	completedSQL, completedArgs, err := database.QB.Select(
		"DATE(updated_at) AS date",
		"COUNT(*)",
	).From("todos").
		Where(sq.Eq{"user_id": userID, "status": TodoStatusDone}).
		Where(sq.GtOrEq{"updated_at": startDate}).
		GroupBy("DATE(updated_at)").
		ToSql()
	if err != nil {
		return nil, err
	}

	completedRows, err := database.GetDB(ctx, r.db).Query(ctx, completedSQL, completedArgs...)
	if err != nil {
		return nil, err
	}
	defer completedRows.Close()

	for completedRows.Next() {
		var dateVal time.Time
		var count int
		if err := completedRows.Scan(&dateVal, &count); err != nil {
			return nil, err
		}
		dateStr := dateVal.Format("2006-01-02")
		if idx, ok := dateMap[dateStr]; ok {
			stats.DailyStats[idx].Completed = count
		}
	}
	if err := completedRows.Err(); err != nil {
		return nil, err
	}

	return &stats, nil
}
