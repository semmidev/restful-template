package delivery

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
)

// ─── Route Registration ─────────────────────────────────────────────────────

func RegisterRoutes(api huma.API, authUsecase *auth.Usecase, todosUsecase *todos.Usecase, log *slog.Logger) {
	RegisterHealthRoutes(api)
	auth.RegisterAuthRoutes(api, authUsecase)
	todos.RegisterTodoRoutes(api, todosUsecase)
}
