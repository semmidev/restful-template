package auth

import (
	"context"
	dbsql "database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *User) error {
	sql, args, err := database.QB.Insert("users").
		Columns("id", "email", "password_hash", "google_id", "active_role", "created_at", "updated_at").
		Values(u.ID, u.Email, u.PasswordHash, u.GoogleID, u.ActiveRole, u.CreatedAt, u.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).ExecContext(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return err
	}

	for _, role := range u.Roles {
		roleSQL, roleArgs, err := database.QB.Insert("user_roles").
			Columns("user_id", "role_name").
			Values(u.ID, role).
			ToSql()
		if err != nil {
			return err
		}
		_, err = database.GetDB(ctx, r.db).ExecContext(ctx, roleSQL, roleArgs...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *userRepository) loadRoles(ctx context.Context, u *User) error {
	sql, args, err := database.QB.Select("role_name").
		From("user_roles").
		Where("user_id = ?", u.ID).
		ToSql()
	if err != nil {
		return err
	}

	var roles []string
	err = database.GetDB(ctx, r.db).SelectContext(ctx, &roles, sql, args...)
	if err != nil {
		return err
	}

	u.Roles = roles
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	sql, args, err := database.QB.Select("id", "email", "password_hash", "google_id", "active_role", "created_at", "updated_at").
		From("users").
		Where("email = ?", email).
		ToSql()
	if err != nil {
		return nil, err
	}

	var u User
	err = database.GetDB(ctx, r.db).GetContext(ctx, &u, sql, args...)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if err := r.loadRoles(ctx, &u); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	sql, args, err := database.QB.Select("id", "email", "password_hash", "google_id", "active_role", "created_at", "updated_at").
		From("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, err
	}

	var u User
	err = database.GetDB(ctx, r.db).GetContext(ctx, &u, sql, args...)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if err := r.loadRoles(ctx, &u); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	sql, args, err := database.QB.Select("id", "email", "password_hash", "google_id", "active_role", "created_at", "updated_at").
		From("users").
		Where("google_id = ?", googleID).
		ToSql()
	if err != nil {
		return nil, err
	}

	var u User
	err = database.GetDB(ctx, r.db).GetContext(ctx, &u, sql, args...)
	if err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if err := r.loadRoles(ctx, &u); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *User) error {
	sql, args, err := database.QB.Update("users").
		Set("email", u.Email).
		Set("password_hash", u.PasswordHash).
		Set("google_id", u.GoogleID).
		Set("active_role", u.ActiveRole).
		Set("updated_at", u.UpdatedAt).
		Where("id = ?", u.ID).
		ToSql()
	if err != nil {
		return err
	}

	res, err := database.GetDB(ctx, r.db).ExecContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	delSQL, delArgs, err := database.QB.Delete("user_roles").
		Where("user_id = ?", u.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = database.GetDB(ctx, r.db).ExecContext(ctx, delSQL, delArgs...)
	if err != nil {
		return err
	}

	for _, role := range u.Roles {
		roleSQL, roleArgs, err := database.QB.Insert("user_roles").
			Columns("user_id", "role_name").
			Values(u.ID, role).
			ToSql()
		if err != nil {
			return err
		}
		_, err = database.GetDB(ctx, r.db).ExecContext(ctx, roleSQL, roleArgs...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := database.QB.Delete("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return err
	}

	res, err := database.GetDB(ctx, r.db).ExecContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
