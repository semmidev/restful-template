package delivery

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/modules/adm/users"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
)

// RegisterRoutes wires all module routes onto the Huma API.
//
// Logging is handled at the middleware layer via canonical wide events;
// individual route registration functions must not receive a logger.
func RegisterRoutes(
	api huma.API,
	cfg config.Config,
	healthCheckers map[string]HealthChecker,
	authService auth.AuthService,
	todosService todos.TodoService,
	usersService users.UserService,
) {
	RegisterHealthRoutes(api, healthCheckers)
	auth.RegisterAuthRoutes(api, authService, cfg)
	todos.RegisterTodoRoutes(api, todosService)
	users.RegisterUserRoutes(api, usersService)
}
