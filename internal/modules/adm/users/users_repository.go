package users

import (
	"context"
	dbsql "database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

var allowedSortCols = map[string]string{
	"created_at":  "created_at",
	"updated_at":  "updated_at",
	"email":       "email",
	"active_role": "active_role",
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *User) error {
	sql, args, err := database.QB.Insert("users").
		Columns("id", "email", "password_hash", "google_id", "active_role", "created_at", "updated_at").
		Values(u.ID, u.Email, u.PasswordHash, nil, u.ActiveRole, u.CreatedAt, u.UpdatedAt).
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

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	sql, args, err := database.QB.Select("id", "email", "password_hash", "active_role", "created_at", "updated_at").
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

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	sql, args, err := database.QB.Select("id", "email", "password_hash", "active_role", "created_at", "updated_at").
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

type userRole struct {
	UserID   uuid.UUID `db:"user_id"`
	RoleName string    `db:"role_name"`
}

func (r *userRepository) List(ctx context.Context, limit, offset int, search string, sortBy, sortDir string) ([]*User, int, error) {
	base := database.QB.Select().From("users")

	if search != "" {
		like := "%" + search + "%"
		base = base.Where(sq.ILike{"email": like})
	}

	col := "created_at"
	if c, ok := allowedSortCols[sortBy]; ok {
		col = c
	}

	dir := "DESC"
	if sortDir == "asc" {
		dir = "ASC"
	}

	sql, args, err := base.
		Columns("id", "email", "password_hash", "active_role", "created_at", "updated_at", "COUNT(*) OVER() AS total_count").
		OrderBy(col + " " + dir).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := database.GetDB(ctx, r.db).QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var users []*User
	var total int
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.ActiveRole, &u.CreatedAt, &u.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(users) == 0 {
		return users, total, nil
	}

	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	roleSQL, roleArgs, err := database.QB.Select("user_id", "role_name").
		From("user_roles").
		Where(sq.Eq{"user_id": userIDs}).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var roles []userRole
	err = database.GetDB(ctx, r.db).SelectContext(ctx, &roles, roleSQL, roleArgs...)
	if err != nil {
		return nil, 0, err
	}

	roleMap := make(map[uuid.UUID][]string)
	for _, r := range roles {
		roleMap[r.UserID] = append(roleMap[r.UserID], r.RoleName)
	}

	for _, u := range users {
		u.Roles = roleMap[u.ID]
		if u.Roles == nil {
			u.Roles = []string{}
		}
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, u *User) error {
	sql, args, err := database.QB.Update("users").
		Set("email", u.Email).
		Set("password_hash", u.PasswordHash).
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
