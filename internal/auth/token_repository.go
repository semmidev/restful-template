package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	sql, args, err := database.QB.Insert("refresh_tokens").
		Columns("token_hash", "user_id", "expires_at").
		Values(tokenHash, userID, expiresAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = database.GetDB(ctx, r.db).ExecContext(ctx, sql, args...)
	return err
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	sql, args, err := database.QB.Delete("refresh_tokens").
		Where("token_hash = ?", tokenHash).
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

func (r *tokenRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	sql, args, err := database.QB.Delete("refresh_tokens").
		Where("expires_at < NOW()").
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := database.GetDB(ctx, r.db).ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
