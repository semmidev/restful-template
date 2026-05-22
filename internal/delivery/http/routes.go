package delivery

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/domain"
)

// ─── Request / Response types ──────────────────────────────────────────────

type RegisterBody struct {
	Email    string `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
	Password string `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
}

type LoginBody = RegisterBody

type RefreshBody struct {
	RefreshToken string `json:"refresh_token" minLength:"1"`
}

type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in" doc:"Unix timestamp when the access token expires"`
}

type TokenResp struct {
	Body struct {
		Data TokenData `json:"data"`
	}
}

type CreateTodoForm struct {
	Title       string        `form:"title" minLength:"1" maxLength:"200"`
	Description string        `form:"description" maxLength:"2000"`
	Cover       huma.FormFile `form:"cover" contentType:"image/*" doc:"Image file to upload as cover"`
}

type UpdateTodoForm struct {
	Title       string        `form:"title" maxLength:"200"`
	Description string        `form:"description" maxLength:"2000"`
	Status      string        `form:"status"`
	Cover       huma.FormFile `form:"cover" contentType:"image/*" doc:"Image file to upload as cover"`
}

type TodoResp struct {
	Body struct {
		Data *domain.Todo `json:"data"`
	}
}

type ListData struct {
	Items   []*domain.Todo `json:"items"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"per_page"`
	Keyword string         `json:"keyword,omitempty" doc:"Active keyword filter (empty if none)"`
	SortBy  string         `json:"sort_by" doc:"Column sorted by"`
	SortDir string         `json:"sort_dir" doc:"Sort direction"`
}

type ListResp struct {
	Body struct {
		Data ListData `json:"data"`
	}
}

type HealthResp struct {
	Body struct {
		Data map[string]string `json:"data"`
	}
}

// ─── Route Registration ─────────────────────────────────────────────────────

func RegisterRoutes(api huma.API, auth domain.AuthUsecase, todos domain.TodoUsecase, log *slog.Logger) {
	// ── Health ──────────────────────────────────────────────────────────────
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health check",
		Tags:        []string{"System"},
	}, func(ctx context.Context, _ *struct{}) (*HealthResp, error) {
		resp := &HealthResp{}
		resp.Body.Data = map[string]string{"status": "ok"}
		return resp, nil
	})

	// ── Auth ────────────────────────────────────────────────────────────────
	huma.Register(api, huma.Operation{
		OperationID: "auth-register",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body RegisterBody }) (*TokenResp, error) {
		pair, err := auth.Register(ctx, domain.RegisterInput{Email: in.Body.Email, Password: in.Body.Password})
		if err != nil {
			return nil, toHumaErr(err)
		}
		return tokenResp(pair), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Login and receive tokens",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body LoginBody }) (*TokenResp, error) {
		pair, err := auth.Login(ctx, domain.LoginInput{Email: in.Body.Email, Password: in.Body.Password})
		if err != nil {
			return nil, toHumaErr(err)
		}
		return tokenResp(pair), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-refresh",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/refresh",
		Summary:     "Refresh access token using a refresh token",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body RefreshBody }) (*TokenResp, error) {
		pair, err := auth.Refresh(ctx, in.Body.RefreshToken)
		if err != nil {
			return nil, toHumaErr(err)
		}
		return tokenResp(pair), nil
	})

	// ── Todos (authenticated) ────────────────────────────────────────────────
	huma.Register(api, huma.Operation{
		OperationID: "list-todos",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos",
		Summary:     "List todos for the authenticated user",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, in *struct {
		Page    int    `query:"page"      default:"1"  minimum:"1"`
		PerPage int    `query:"per_page"  default:"20" minimum:"1" maximum:"100"`
		Status  string `query:"status"  enum:"pending,in_progress,done," doc:"Filter by status (empty = all)"`
		Keyword string `query:"q"       maxLength:"100"              doc:"Case-insensitive substring search on title and description"`
		SortBy  string `query:"sort_by" default:"created_at" enum:"created_at,updated_at,title,status" doc:"Column to sort by"`
		SortDir string `query:"sort_dir" default:"desc" enum:"asc,desc" doc:"Sort direction"`
	}) (*ListResp, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, toHumaErr(err)
		}
		offset := (in.Page - 1) * in.PerPage

		items, total, err := todos.List(ctx, domain.ListTodosQuery{
			UserID:  userID,
			Limit:   in.PerPage,
			Offset:  offset,
			Status:  in.Status,
			Keyword: in.Keyword,
			SortBy:  in.SortBy,
			SortDir: in.SortDir,
		})
		if err != nil {
			return nil, toHumaErr(err)
		}
		resp := &ListResp{}
		resp.Body.Data.Items = items
		resp.Body.Data.Total = total
		resp.Body.Data.Page = in.Page
		resp.Body.Data.PerPage = in.PerPage
		resp.Body.Data.Keyword = in.Keyword
		resp.Body.Data.SortBy = in.SortBy
		resp.Body.Data.SortDir = in.SortDir
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-todo",
		Method:        http.MethodPost,
		Path:          "/api/v1/todos",
		Summary:       "Create a new todo",
		Tags:          []string{"Todos"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *struct {
		RawBody huma.MultipartFormFiles[CreateTodoForm]
	}) (*TodoResp, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, toHumaErr(err)
		}

		data := in.RawBody.Data()
		var coverBase64 *string
		if data.Cover.IsSet {
			b, err := io.ReadAll(data.Cover)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read cover image")
			}
			encoded := fmt.Sprintf("data:%s;base64,%s", data.Cover.ContentType, base64.StdEncoding.EncodeToString(b))
			coverBase64 = &encoded
		}

		var desc *string
		if data.Description != "" {
			desc = &data.Description
		}

		t, err := todos.Create(ctx, domain.CreateTodoInput{
			UserID:      userID,
			Title:       data.Title,
			Description: desc,
			Cover:       coverBase64,
		})
		if err != nil {
			return nil, toHumaErr(err)
		}
		resp := &TodoResp{}
		resp.Body.Data = t
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-todo",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Get a single todo by ID",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, in *struct {
		ID uuid.UUID `path:"id" doc:"Todo UUID"`
	}) (*TodoResp, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, toHumaErr(err)
		}
		t, err := todos.Get(ctx, userID, in.ID)
		if err != nil {
			return nil, toHumaErr(err)
		}
		resp := &TodoResp{}
		resp.Body.Data = t
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-todo",
		Method:      http.MethodPatch,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Update a todo",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, in *struct {
		ID      uuid.UUID `path:"id"`
		RawBody huma.MultipartFormFiles[UpdateTodoForm]
	}) (*TodoResp, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, toHumaErr(err)
		}

		data := in.RawBody.Data()

		var title *string
		if _, ok := in.RawBody.Form.Value["title"]; ok {
			title = &data.Title
		}

		var desc *string
		if _, ok := in.RawBody.Form.Value["description"]; ok {
			desc = &data.Description
		}

		var status *domain.TodoStatus
		if _, ok := in.RawBody.Form.Value["status"]; ok && data.Status != "" {
			st := domain.TodoStatus(data.Status)
			status = &st
		}

		var coverBase64 *string
		if data.Cover.IsSet {
			b, err := io.ReadAll(data.Cover)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read cover image")
			}
			encoded := fmt.Sprintf("data:%s;base64,%s", data.Cover.ContentType, base64.StdEncoding.EncodeToString(b))
			coverBase64 = &encoded
		}

		t, err := todos.Update(ctx, domain.UpdateTodoInput{
			ID:          in.ID,
			UserID:      userID,
			Title:       title,
			Description: desc,
			Cover:       coverBase64,
			Status:      status,
		})
		if err != nil {
			return nil, toHumaErr(err)
		}
		resp := &TodoResp{}
		resp.Body.Data = t
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-todo",
		Method:        http.MethodDelete,
		Path:          "/api/v1/todos/{id}",
		Summary:       "Delete a todo",
		Tags:          []string{"Todos"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, in *struct {
		ID uuid.UUID `path:"id"`
	}) (*struct{}, error) {
		userID, err := extractUserID(ctx)
		if err != nil {
			return nil, toHumaErr(err)
		}
		if err := todos.Delete(ctx, userID, in.ID); err != nil {
			return nil, toHumaErr(err)
		}
		return &struct{}{}, nil
	})
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func tokenResp(pair domain.TokenPair) *TokenResp {
	r := &TokenResp{}
	r.Body.Data.AccessToken = pair.AccessToken
	r.Body.Data.RefreshToken = pair.RefreshToken
	r.Body.Data.ExpiresIn = pair.ExpiresIn
	return r
}

func extractUserID(ctx context.Context) (uuid.UUID, error) {
	raw := GetUserID(ctx)
	if raw == "" {
		return uuid.Nil, domain.ErrUnauthorized
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return id, nil
}

// toHumaErr maps domain errors to RFC 9457 problem+json Huma errors.
func toHumaErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return huma.Error404NotFound("resource not found")
	case errors.Is(err, domain.ErrConflict):
		return huma.Error409Conflict("resource already exists")
	case errors.Is(err, domain.ErrUnauthorized):
		return huma.Error401Unauthorized("unauthorized")
	case errors.Is(err, domain.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, domain.ErrInvalidInput):
		return huma.Error400BadRequest("invalid input")
	default:
		return huma.Error500InternalServerError("internal server error")
	}
}
