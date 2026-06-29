package todo

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type todoHandler struct {
	todos TodoService
}

func RegisterTodoRoutes(api huma.API, todos TodoService) {
	h := &todoHandler{todos: todos}

	huma.Register(api, huma.Operation{
		OperationID: "list-todos",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos",
		Summary:     "List todos for the authenticated user",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleList)

	huma.Register(api, huma.Operation{
		OperationID:   "create-todo",
		Method:        http.MethodPost,
		Path:          "/api/v1/todos",
		Summary:       "Create a new todo",
		Tags:          []string{"Todos"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.handleCreate)

	huma.Register(api, huma.Operation{
		OperationID: "get-todo",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Get a single todo by ID",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGet)

	huma.Register(api, huma.Operation{
		OperationID: "update-todo",
		Method:      http.MethodPatch,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Update a todo",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleUpdate)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-todo",
		Method:        http.MethodDelete,
		Path:          "/api/v1/todos/{id}",
		Summary:       "Delete a todo",
		Tags:          []string{"Todos"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.handleDelete)

	huma.Register(api, huma.Operation{
		OperationID: "restore-todo",
		Method:      http.MethodPost,
		Path:        "/api/v1/todos/{id}/restore",
		Summary:     "Restore a soft-deleted todo",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleRestore)

	huma.Register(api, huma.Operation{
		OperationID: "get-todo-stats",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos/stats",
		Summary:     "Get todo statistics for the authenticated user",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.handleStats)
}
