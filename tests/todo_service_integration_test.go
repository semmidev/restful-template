package tests

import (
	"context"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/domain"
	"github.com/semmidev/restful-template/internal/infrastructure/repository/postgres"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
	"github.com/semmidev/restful-template/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func insertDummyUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool) uuid.UUID {
	userID := uuidgen.New()
	sql, args, err := psql.Insert("users").
		Columns("id", "email", "password_hash").
		Values(userID, userID.String()+"@example.com", "hash").
		ToSql()
	require.NoError(t, err)

	_, err = pool.Exec(ctx, sql, args...)
	require.NoError(t, err)
	return userID
}

func TestTodoService_Integration(t *testing.T) {
	pool, cleanup := SetupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	repo := postgres.NewTodoRepository(pool)
	svc := usecase.NewTodo(repo, nil, nil)

	userID := insertDummyUser(t, ctx, pool)

	t.Run("Create Todo", func(t *testing.T) {
		in := domain.CreateTodoInput{
			UserID: userID,
			Title:  "Integration Test Todo",
		}
		todo, err := svc.Create(ctx, in)
		assert.NoError(t, err)
		assert.NotNil(t, todo)
		assert.Equal(t, "Integration Test Todo", todo.Title)
		assert.Equal(t, domain.TodoPending, todo.Status)
	})

	t.Run("Create validation failure", func(t *testing.T) {
		in := domain.CreateTodoInput{
			UserID: userID,
			Title:  "",
		}
		todo, err := svc.Create(ctx, in)
		assert.Error(t, err)
		assert.Nil(t, todo)
	})

	t.Run("Update Todo", func(t *testing.T) {
		in := domain.CreateTodoInput{
			UserID: userID,
			Title:  "Old Title",
		}
		todo, err := svc.Create(ctx, in)
		require.NoError(t, err)

		newTitle := "New Title"
		updated, err := svc.Update(ctx, domain.UpdateTodoInput{
			ID:     todo.ID,
			UserID: userID,
			Title:  &newTitle,
		})
		assert.NoError(t, err)
		assert.Equal(t, "New Title", updated.Title)
	})

	t.Run("List Todos", func(t *testing.T) {
		// Insert another user to test isolation
		otherUserID := insertDummyUser(t, ctx, pool)

		// Create two for primary user
		_, err := svc.Create(ctx, domain.CreateTodoInput{UserID: userID, Title: "List item 1"})
		require.NoError(t, err)
		_, err = svc.Create(ctx, domain.CreateTodoInput{UserID: userID, Title: "List item 2"})
		require.NoError(t, err)

		// Create one for other user
		_, err = svc.Create(ctx, domain.CreateTodoInput{UserID: otherUserID, Title: "Other user item"})
		require.NoError(t, err)

		items, total, err := svc.List(ctx, domain.ListTodosQuery{
			UserID: userID,
			Limit:  10,
		})
		assert.NoError(t, err)

		// The total might be more than 2 if previous tests also created todos for userID, but
		// the ones for otherUserID should NOT be there.
		assert.GreaterOrEqual(t, total, 2)

		// Ensure none of the returned items belong to otherUserID
		for _, item := range items {
			assert.Equal(t, userID, item.UserID)
		}
	})

	t.Run("Delete Todo", func(t *testing.T) {
		in := domain.CreateTodoInput{
			UserID: userID,
			Title:  "To be deleted",
		}
		todo, err := svc.Create(ctx, in)
		require.NoError(t, err)

		err = svc.Delete(ctx, userID, todo.ID)
		assert.NoError(t, err)

		_, err = svc.Get(ctx, userID, todo.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
