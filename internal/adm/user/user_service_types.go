package user

import (
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type CreateUserInput struct {
	Email      string   `json:"email"`
	Password   string   `json:"password"`
	ActiveRole string   `json:"active_role"`
	Roles      []string `json:"roles"`
}

func (in *CreateUserInput) Validate() error {
	if in.Email == "" || len(in.Password) < 8 {
		return apperrors.NewInvalidInput("Email is required, password must be at least 8 characters", apperrors.ErrInvalidInput)
	}
	if len(in.Roles) == 0 {
		return apperrors.NewInvalidInput("User must have at least one role", apperrors.ErrInvalidInput)
	}
	activeRoleValid := false
	for _, r := range in.Roles {
		if r == in.ActiveRole {
			activeRoleValid = true
			break
		}
	}
	if !activeRoleValid {
		return apperrors.NewInvalidInput("Active role must be one of the assigned roles", apperrors.ErrInvalidInput)
	}
	return nil
}

type UpdateUserInput struct {
	Email      *string  `json:"email,omitempty"`
	Password   *string  `json:"password,omitempty"`
	ActiveRole *string  `json:"active_role,omitempty"`
	Roles      []string `json:"roles,omitempty"`
}

func (in *UpdateUserInput) Validate(currentActiveRole string, currentRoles []string) error {
	if in.Email != nil && *in.Email == "" {
		return apperrors.NewInvalidInput("Email cannot be empty", apperrors.ErrInvalidInput)
	}
	if in.Password != nil && len(*in.Password) < 8 {
		return apperrors.NewInvalidInput("Password must be at least 8 characters", apperrors.ErrInvalidInput)
	}

	targetRoles := currentRoles
	if in.Roles != nil {
		targetRoles = in.Roles
	}

	targetActiveRole := currentActiveRole
	if in.ActiveRole != nil {
		targetActiveRole = *in.ActiveRole
	}

	if len(targetRoles) == 0 {
		return apperrors.NewInvalidInput("User must have at least one role", apperrors.ErrInvalidInput)
	}

	activeRoleValid := false
	for _, r := range targetRoles {
		if r == targetActiveRole {
			activeRoleValid = true
			break
		}
	}
	if !activeRoleValid {
		return apperrors.NewInvalidInput("Active role must be one of the assigned roles", apperrors.ErrInvalidInput)
	}

	return nil
}
