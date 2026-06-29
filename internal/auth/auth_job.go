package auth

import (
	"context"
	"log/slog"
)

type AuthJob struct {
	repo   TokenRepository
	logger *slog.Logger
}

func NewAuthJob(repo TokenRepository, logger *slog.Logger) *AuthJob {
	return &AuthJob{repo: repo, logger: logger}
}

// CleanupExpiredTokens deletes expired refresh tokens from the database.
func (j *AuthJob) CleanupExpiredTokens() {
	ctx := context.Background()
	n, err := j.repo.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		j.logger.Error("refresh token cleanup failed", "err", err)
		return
	}
	if n > 0 {
		j.logger.Info("cleaned up expired refresh tokens", "count", n)
	}
}
