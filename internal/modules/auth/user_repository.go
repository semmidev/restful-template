package auth

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

type userRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository { return &userRepository{db} }

func (r *userRepository) Create(ctx context.Context, u *User) error {
	sql, args, err := psql.Insert("users").
		Columns("id", "email", "password_hash", "created_at", "updated_at").
		Values(u.ID, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	sql, args, err := psql.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := database.GetDB(ctx, r.db).QueryRow(ctx, sql, args...)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	sql, args, err := psql.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := database.GetDB(ctx, r.db).QueryRow(ctx, sql, args...)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := psql.Delete("users").
		Where(sq.Eq{"id": id}).
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
