package http

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/adm/user"
	"github.com/semmidev/restful-template/internal/auth"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/todo"
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
	todoService todo.TodoService,
	usersService user.UserService,
) {
	RegisterHealthRoutes(api, healthCheckers)
	auth.RegisterAuthRoutes(api, authService, cfg)
	todo.RegisterTodoRoutes(api, todoService)
	user.RegisterUserRoutes(api, usersService)
}
