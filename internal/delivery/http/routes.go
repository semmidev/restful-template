package delivery

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
)

// ─── Route Registration ─────────────────────────────────────────────────────

// RegisterRoutes wires all module routes onto the Huma API.
//
// point 16: removed the unused `log *slog.Logger` parameter.
// Logging is handled at the middleware layer via canonical wide events;
// individual route registrations don't need a logger reference.
func RegisterRoutes(
	api huma.API,
	healthCheckers map[string]HealthChecker,
	authService auth.AuthService,
	todosService todos.TodoService,
) {
	RegisterHealthRoutes(api, healthCheckers)
	auth.RegisterAuthRoutes(api, authService)
	todos.RegisterTodoRoutes(api, todosService)
}
