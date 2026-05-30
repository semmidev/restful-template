package jobs

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CleanupExpiredTokens deletes expired refresh tokens from the database.
func CleanupExpiredTokens(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) {
	tag, err := pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE expires_at < NOW()")
	if err != nil {
		logger.Error("refresh token cleanup failed", "err", err)
		return
	}
	if n := tag.RowsAffected(); n > 0 {
		logger.Info("cleaned up expired refresh tokens", "count", n)
	}
}
