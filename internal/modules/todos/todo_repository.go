package todos

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// TodoRepository is the postgres-backed driven adapter for todos.
type todoRepository struct{ db *pgxpool.Pool }

func NewTodoRepository(db *pgxpool.Pool) TodoRepository { return &todoRepository{db} }

func (r *todoRepository) Create(ctx context.Context, t *Todo) error {
	sql, args, err := psql.Insert("todos").
		Columns("id", "user_id", "title", "description", "cover", "status", "created_at", "updated_at").
		Values(t.ID, t.UserID, t.Title, t.Description, t.Cover, t.Status, t.CreatedAt, t.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *todoRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Todo, error) {
	sql, args, err := psql.Select("id", "user_id", "title", "description", "cover", "status", "created_at", "updated_at").
		From("todos").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := database.GetDB(ctx, r.db).QueryRow(ctx, sql, args...)
	var t Todo
	if err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Cover, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// ListByUser returns paginated todos for a user with optional status and keyword filters.
func (r *todoRepository) ListByUser(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error) {

	base := psql.Select().From("todos").Where(sq.Eq{"user_id": q.UserID})

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

	sortBy := "created_at"
	if q.SortBy == "title" || q.SortBy == "updated_at" || q.SortBy == "status" {
		sortBy = q.SortBy
	}

	sortDir := "DESC"
	if q.SortDir == "asc" {
		sortDir = "ASC"
	}

	// 1. Data query
	dataQuery := base.
		Columns("id", "user_id", "title", "description", "cover", "status", "created_at", "updated_at").
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

	out := make([]*Todo, 0)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Cover, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// 2. Count query
	countSQL, countArgs, err := base.Columns("COUNT(*)").ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := database.GetDB(ctx, r.db).QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *todoRepository) Update(ctx context.Context, t *Todo) error {
	sql, args, err := psql.Update("todos").
		Set("title", t.Title).
		Set("description", t.Description).
		Set("cover", t.Cover).
		Set("status", t.Status).
		Set("updated_at", t.UpdatedAt).
		Where(sq.Eq{"id": t.ID, "user_id": t.UserID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *todoRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	sql, args, err := psql.Delete("todos").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *todoRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	sql, args, err := psql.Delete("todos").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}
