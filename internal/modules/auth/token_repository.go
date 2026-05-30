package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) TokenRepository {
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

	_, err = database.GetDB(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	sql, args, err := database.QB.Delete("refresh_tokens").
		Where("token_hash = ?", tokenHash).
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

func (r *tokenRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	tag, err := database.GetDB(ctx, r.db).Exec(ctx, "DELETE FROM refresh_tokens WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
