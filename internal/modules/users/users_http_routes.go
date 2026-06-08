package users

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterUserRoutes(api huma.API, service UserService) {
	h := &userHandler{service: service}

	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/api/v1/admin/users",
		Summary:     "List all registered users (Admin only)",
		Tags:        []string{"User Management"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleList)

	huma.Register(api, huma.Operation{
		OperationID: "get-user",
		Method:      http.MethodGet,
		Path:        "/api/v1/admin/users/{id}",
		Summary:     "Retrieve user by ID (Admin only)",
		Tags:        []string{"User Management"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGet)

	huma.Register(api, huma.Operation{
		OperationID:   "create-user",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/users",
		Summary:       "Create a new user (Admin only)",
		Tags:          []string{"User Management"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.handleCreate)

	huma.Register(api, huma.Operation{
		OperationID: "update-user",
		Method:      http.MethodPatch,
		Path:        "/api/v1/admin/users/{id}",
		Summary:     "Update a user's details (Admin only)",
		Tags:        []string{"User Management"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleUpdate)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-user",
		Method:        http.MethodDelete,
		Path:          "/api/v1/admin/users/{id}",
		Summary:       "Delete a user (Admin only)",
		Tags:          []string{"User Management"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.handleDelete)
}
