package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/domain"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock repository ────────────────────────────────────────────────────────

type mockTodoRepo struct{ mock.Mock }

func (m *mockTodoRepo) Create(ctx context.Context, t *domain.Todo) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTodoRepo) FindByID(ctx context.Context, u, i uuid.UUID) (*domain.Todo, error) {
	a := m.Called(ctx, u, i)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Todo), a.Error(1)
}
func (m *mockTodoRepo) ListByUser(ctx context.Context, q domain.ListTodosQuery) ([]*domain.Todo, int, error) {
	a := m.Called(ctx, q)
	return a.Get(0).([]*domain.Todo), a.Int(1), a.Error(2)
}
func (m *mockTodoRepo) Update(ctx context.Context, t *domain.Todo) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTodoRepo) Delete(ctx context.Context, u, i uuid.UUID) error {
	return m.Called(ctx, u, i).Error(0)
}

// ─── Table-driven tests ─────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	tests := []struct {
		name      string
		in        domain.CreateTodoInput
		repoErr   error
		wantErr   bool
		wantTitle string
	}{
		{
			name:      "success — generates UUID v7",
			in:        domain.CreateTodoInput{UserID: uuidgen.New(), Title: "Test todo"},
			repoErr:   nil,
			wantTitle: "Test todo",
		},
		{
			name:    "empty title fails domain validation",
			in:      domain.CreateTodoInput{UserID: uuidgen.New(), Title: ""},
			wantErr: true,
		},
		{
			name:    "title too long fails domain validation",
			in:      domain.CreateTodoInput{UserID: uuidgen.New(), Title: string(make([]byte, 201))},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockTodoRepo)
			svc := NewTodoService(repo, nil, nil)

			if !tc.wantErr {
				repo.On("Create", mock.Anything, mock.Anything).Return(tc.repoErr)
			}

			todo, err := svc.Create(context.Background(), tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, todo)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTitle, todo.Title)
				assert.Equal(t, domain.TodoPending, todo.Status)
				// UUID v7 version bit = 7
				assert.Equal(t, 7, int(todo.ID.Version()))
			}
		})
	}
}

func TestList(t *testing.T) {
	userID := uuidgen.New()

	tests := []struct {
		name      string
		query     domain.ListTodosQuery
		mockItems []*domain.Todo
		mockTotal int
		wantLimit int
	}{
		{
			name:      "zero limit gets clamped to default 20",
			query:     domain.ListTodosQuery{UserID: userID, Limit: 0},
			mockItems: []*domain.Todo{},
			mockTotal: 0,
			wantLimit: 20,
		},
		{
			name:      "custom limit respected",
			query:     domain.ListTodosQuery{UserID: userID, Limit: 5},
			mockItems: []*domain.Todo{{Title: "a"}, {Title: "b"}},
			mockTotal: 2,
			wantLimit: 5,
		},
		{
			name:      "over-limit clamped to 20",
			query:     domain.ListTodosQuery{UserID: userID, Limit: 200},
			mockItems: []*domain.Todo{},
			mockTotal: 0,
			wantLimit: 20,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockTodoRepo)
			svc := NewTodoService(repo, nil, nil)

			expectedQuery := tc.query
			expectedQuery.Limit = tc.wantLimit
			expectedQuery.SortBy = "created_at"
			expectedQuery.SortDir = "desc"
			repo.On("ListByUser", mock.Anything, expectedQuery).
				Return(tc.mockItems, tc.mockTotal, nil)

			items, total, err := svc.List(context.Background(), tc.query)
			assert.NoError(t, err)
			assert.Equal(t, tc.mockTotal, total)
			assert.Equal(t, tc.mockItems, items)
		})
	}
}

func TestDelete(t *testing.T) {
	repo := new(mockTodoRepo)
	svc := NewTodoService(repo, nil, nil)
	userID := uuidgen.New()
	id := uuidgen.New()

	repo.On("Delete", mock.Anything, userID, id).Return(nil)

	err := svc.Delete(context.Background(), userID, id)
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	repo := new(mockTodoRepo)
	svc := NewTodoService(repo, nil, nil)
	userID := uuidgen.New()
	id := uuidgen.New()

	existing := &domain.Todo{ID: id, UserID: userID, Title: "Old", Status: domain.TodoPending}
	repo.On("FindByID", mock.Anything, userID, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	newTitle := "New Title"
	updated, err := svc.Update(context.Background(), domain.UpdateTodoInput{
		ID:     id,
		UserID: userID,
		Title:  &newTitle,
	})
	assert.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
}
