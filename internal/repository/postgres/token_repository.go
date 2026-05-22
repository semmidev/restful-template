package postgres

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/domain"
)

type TokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) StoreRefreshToken(ctx context.Context, tokenHash string, userID uuid.UUID, expiresAt time.Time) error {
	sql, args, err := psql.Insert("refresh_tokens").
		Columns("token_hash", "user_id", "expires_at").
		Values(tokenHash, userID, expiresAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = getDb(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	sql, args, err := psql.Delete("refresh_tokens").
		Where(sq.Eq{"token_hash": tokenHash}).
		ToSql()
	if err != nil {
		return err
	}

	res, err := getDb(ctx, r.db).Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
