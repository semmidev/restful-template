package delivery

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
)

// ─── Route Registration ─────────────────────────────────────────────────────

func RegisterRoutes(api huma.API, authService auth.AuthService, todosService todos.TodoService, log *slog.Logger) {
	RegisterHealthRoutes(api)
	auth.RegisterAuthRoutes(api, authService)
	todos.RegisterTodoRoutes(api, todosService)
}
