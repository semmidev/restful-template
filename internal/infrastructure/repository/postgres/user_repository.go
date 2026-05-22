package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/domain"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db} }

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	sql, args, err := psql.Insert("users").
		Columns("id", "email", "password_hash", "created_at", "updated_at").
		Values(u.ID, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = getDb(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	sql, args, err := psql.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := getDb(ctx, r.db).QueryRow(ctx, sql, args...)
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	sql, args, err := psql.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := getDb(ctx, r.db).QueryRow(ctx, sql, args...)
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
