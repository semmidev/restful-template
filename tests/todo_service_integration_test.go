package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	delivery "github.com/semmidev/restful-template/internal/delivery/http"
	"github.com/semmidev/restful-template/internal/infrastructure/jwt"
	pgRepo "github.com/semmidev/restful-template/internal/infrastructure/repository/postgres"
	"github.com/semmidev/restful-template/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key-minimum-32-bytes!!"

// newTestAPI wires up the full Huma API (routes + auth middleware) backed by
// the given pool, without Redis or observability dependencies.
func newTestAPI(pool *pgxpool.Pool) huma.API {
	todoRepo := pgRepo.NewTodoRepository(pool)
	userRepo := pgRepo.NewUserRepository(pool)
	tokenRepo := pgRepo.NewTokenRepository(pool)

	tokenSvc := jwt.NewJWTService(testJWTSecret, 15*time.Minute, 7*24*time.Hour)

	authSvc := usecase.NewAuth(userRepo, tokenSvc, tokenRepo, nil)
	todoSvc := usecase.NewTodo(todoRepo, nil, nil)

	r := chi.NewRouter()
	humaConfig := huma.DefaultConfig("Todo API Test", "0.0.0")
	humaConfig.Components = &huma.Components{
		SecuritySchemes: map[string]*huma.SecurityScheme{
			"bearerAuth": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
		},
	}
	api := humachi.New(r, humaConfig)
	api.UseMiddleware(delivery.AuthMiddleware(api, tokenSvc))
	delivery.RegisterRoutes(api, authSvc, todoSvc, nil)
	return api
}

// registerAndLogin calls the register endpoint and returns the access token.
func registerAndLogin(t *testing.T, api huma.API, email, password string) string {
	t.Helper()
	body := map[string]string{"email": email, "password": password}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "register failed: %s", w.Body.String())

	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp.Data.AccessToken
}

// doRequest is a helper that sends an authenticated request and returns the recorder.
func doRequest(api huma.API, method, path, token string, body []byte, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	return w
}

// buildMultipartBody builds a simple multipart/form-data body with text fields only.
func buildMultipartBody(fields map[string]string) ([]byte, string) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	mw.Close()
	return buf.Bytes(), mw.FormDataContentType()
}

// ─── Tests ──────────────────────────────────────────────────────────────────

func TestTodoHTTP_Integration(t *testing.T) {
	pool, cleanup := SetupTestDatabase(t)
	defer cleanup()

	api := newTestAPI(pool)

	// Each sub-test gets its own isolated user so they don't interfere.
	email := fmt.Sprintf("test-%s@example.com", uuid.New().String())
	token := registerAndLogin(t, api, email, "password123")

	var createdID string

	t.Run("POST /api/v1/todos – success", func(t *testing.T) {
		body, ct := buildMultipartBody(map[string]string{"title": "Integration HTTP Todo"})
		w := doRequest(api, http.MethodPost, "/api/v1/todos", token, body, ct)

		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		var resp struct {
			Data struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Integration HTTP Todo", resp.Data.Title)
		assert.NotEmpty(t, resp.Data.ID)

		createdID = resp.Data.ID
	})

	t.Run("POST /api/v1/todos – missing title returns 422", func(t *testing.T) {
		body, ct := buildMultipartBody(map[string]string{"title": ""})
		w := doRequest(api, http.MethodPost, "/api/v1/todos", token, body, ct)
		// Huma validates minLength:1 -> 422 Unprocessable Entity
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	})

	t.Run("GET /api/v1/todos – returns list", func(t *testing.T) {
		w := doRequest(api, http.MethodGet, "/api/v1/todos", token, nil, "")
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp struct {
			Data struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
				Total int `json:"total"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.GreaterOrEqual(t, resp.Data.Total, 1)
	})

	t.Run("GET /api/v1/todos/{id} – returns single todo", func(t *testing.T) {
		require.NotEmpty(t, createdID, "depends on create test")
		w := doRequest(api, http.MethodGet, "/api/v1/todos/"+createdID, token, nil, "")
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp struct {
			Data struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, createdID, resp.Data.ID)
		assert.Equal(t, "Integration HTTP Todo", resp.Data.Title)
	})

	t.Run("PATCH /api/v1/todos/{id} – updates title", func(t *testing.T) {
		require.NotEmpty(t, createdID, "depends on create test")
		body, ct := buildMultipartBody(map[string]string{"title": "Updated Title"})
		w := doRequest(api, http.MethodPatch, "/api/v1/todos/"+createdID, token, body, ct)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Updated Title", resp.Data.Title)
	})

	t.Run("DELETE /api/v1/todos/{id} – removes todo", func(t *testing.T) {
		require.NotEmpty(t, createdID, "depends on create test")
		w := doRequest(api, http.MethodDelete, "/api/v1/todos/"+createdID, token, nil, "")
		assert.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	})

	t.Run("GET /api/v1/todos/{id} – returns 404 after delete", func(t *testing.T) {
		require.NotEmpty(t, createdID, "depends on delete test")
		w := doRequest(api, http.MethodGet, "/api/v1/todos/"+createdID, token, nil, "")
		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("GET /api/v1/todos – unauthenticated returns 401", func(t *testing.T) {
		w := doRequest(api, http.MethodGet, "/api/v1/todos", "", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())
	})

	t.Run("GET /api/v1/todos/{id} – other user cannot access todo", func(t *testing.T) {
		// Create a second user and their own todo
		email2 := fmt.Sprintf("other-%s@example.com", uuid.New().String())
		token2 := registerAndLogin(t, api, email2, "password123")

		body, ct := buildMultipartBody(map[string]string{"title": "Other user's todo"})
		wCreate := doRequest(api, http.MethodPost, "/api/v1/todos", token2, body, ct)
		require.Equal(t, http.StatusCreated, wCreate.Code)

		var createResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(wCreate.Body.Bytes(), &createResp))

		// user1 tries to access user2's todo — must get 404 (not found for that user)
		w := doRequest(api, http.MethodGet, "/api/v1/todos/"+createResp.Data.ID, token, nil, "")
		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}

// Ensure humatest is imported to satisfy the go.mod dependency.
var _ = humatest.New
