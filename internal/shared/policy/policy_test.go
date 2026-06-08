package policy

import (
	"context"
	"testing"
)

func TestPolicyEvaluator(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx)
	if err != nil {
		t.Fatalf("failed to initialize evaluator: %v", err)
	}

	tests := []struct {
		name    string
		input   Input
		allowed bool
	}{
		{
			name: "Admin can do everything",
			input: Input{
				UserID:          "user-1",
				ActiveRole:      "admin",
				Roles:           []string{"admin", "user"},
				Action:          "todo:delete",
				ResourceOwnerID: "user-2",
			},
			allowed: true,
		},
		{
			name: "User can delete their own todo",
			input: Input{
				UserID:          "user-1",
				ActiveRole:      "user",
				Roles:           []string{"user"},
				Action:          "todo:delete",
				ResourceOwnerID: "user-1",
			},
			allowed: true,
		},
		{
			name: "User cannot delete others' todos",
			input: Input{
				UserID:          "user-1",
				ActiveRole:      "user",
				Roles:           []string{"user"},
				Action:          "todo:delete",
				ResourceOwnerID: "user-2",
			},
			allowed: false,
		},
		{
			name: "User can create todo",
			input: Input{
				UserID:     "user-1",
				ActiveRole: "user",
				Roles:      []string{"user"},
				Action:     "todo:create",
			},
			allowed: true,
		},
		{
			name: "Admin can list users",
			input: Input{
				UserID:     "user-1",
				ActiveRole: "admin",
				Roles:      []string{"admin"},
				Action:     "user:list",
			},
			allowed: true,
		},
		{
			name: "Standard user cannot list users",
			input: Input{
				UserID:     "user-1",
				ActiveRole: "user",
				Roles:      []string{"user"},
				Action:     "user:list",
			},
			allowed: false,
		},
		{
			name: "Standard user cannot create user",
			input: Input{
				UserID:     "user-1",
				ActiveRole: "user",
				Roles:      []string{"user"},
				Action:     "user:create",
			},
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := evaluator.IsAllowed(ctx, tt.input)
			if err != nil {
				t.Fatalf("IsAllowed returned error: %v", err)
			}
			if res != tt.allowed {
				t.Errorf("expected allowed = %v, got %v", tt.allowed, res)
			}
		})
	}
}

func TestGetRolePermissions(t *testing.T) {
	ctx := context.Background()
	err := Init(ctx)
	if err != nil {
		t.Fatalf("failed to initialize evaluator: %v", err)
	}

	adminPerms, err := GetRolePermissions(ctx, "admin")
	if err != nil {
		t.Fatalf("failed to get admin permissions: %v", err)
	}
	found := false
	for _, p := range adminPerms {
		if p == "user:create" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected admin to have user:create permission")
	}

	userPerms, err := GetRolePermissions(ctx, "user")
	if err != nil {
		t.Fatalf("failed to get user permissions: %v", err)
	}
	for _, p := range userPerms {
		if p == "user:create" {
			t.Errorf("expected user not to have user:create permission")
		}
	}
}
