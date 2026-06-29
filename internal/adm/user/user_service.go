package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/semmidev/restful-template/internal/shared/password"
)

type Service struct {
	repo      UserRepository
	txManager database.TxManager
	tracer    observability.Tracer
}

func NewUserService(repo UserRepository, txManager database.TxManager, tracer observability.Tracer) UserService {
	return &Service{
		repo:      repo,
		txManager: txManager,
		tracer:    tracer,
	}
}

func (s *Service) Create(ctx context.Context, in CreateUserInput) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "users.Create")
	defer span.End()

	if err := in.Validate(); err != nil {
		return nil, err
	}

	passHash, err := password.Hash(in.Password)
	if err != nil {
		return nil, apperrors.NewInternal("Failed to process user creation", err)
	}

	u := &User{
		ID:           uuid.New(),
		Email:        in.Email,
		PasswordHash: &passHash,
		ActiveRole:   in.ActiveRole,
		Roles:        in.Roles,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		return s.repo.Create(txCtx, u)
	})

	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, apperrors.NewConflict("Email is already registered", err)
		}
		return nil, apperrors.NewInternal("Failed to create user", err)
	}

	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "users.GetByID")
	defer span.End()

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NewNotFound("User not found", err)
		}
		return nil, apperrors.NewInternal("Failed to retrieve user", err)
	}
	return u, nil
}

func (s *Service) List(ctx context.Context, page, perPage int, search string, sortBy, sortDir string) ([]*User, int, error) {
	ctx, span := s.tracer.Start(ctx, "users.List")
	defer span.End()

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	items, total, err := s.repo.List(ctx, perPage, offset, search, sortBy, sortDir)
	if err != nil {
		return nil, 0, apperrors.NewInternal("Failed to list users", err)
	}

	return items, total, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "users.Update")
	defer span.End()

	var u *User
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		var txErr error
		u, txErr = s.repo.GetByID(txCtx, id)
		if txErr != nil {
			return txErr
		}

		if txErr = in.Validate(u.ActiveRole, u.Roles); txErr != nil {
			return txErr
		}

		if in.Email != nil {
			u.Email = *in.Email
		}
		if in.Password != nil {
			hash, txErr := password.Hash(*in.Password)
			if txErr != nil {
				return apperrors.NewInternal("Failed to hash updated password", txErr)
			}
			u.PasswordHash = &hash
		}
		if in.ActiveRole != nil {
			u.ActiveRole = *in.ActiveRole
		}
		if in.Roles != nil {
			u.Roles = in.Roles
		}
		u.UpdatedAt = time.Now()

		return s.repo.Update(txCtx, u)
	})

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NewNotFound("User not found", err)
		}
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, apperrors.NewConflict("Email is already registered", err)
		}
		var safeErr *apperrors.SafeError
		if errors.As(err, &safeErr) {
			return nil, safeErr
		}
		return nil, apperrors.NewInternal("Failed to update user", err)
	}

	return u, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "users.Delete")
	defer span.End()

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id)
	})

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.NewNotFound("User not found", err)
		}
		return apperrors.NewInternal("Failed to delete user", err)
	}

	return nil
}
