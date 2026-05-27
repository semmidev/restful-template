package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/semmidev/restful-template/internal/app"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestInfrastructure(t *testing.T) (pgDSN string, redisDSN string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	// PostgreSQL
	dbContainer, err := postgres.Run(ctx,
		"docker.io/postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	pgDSN, err = dbContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	// Redis
	redisContainer, err := redis.Run(ctx,
		"docker.io/redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	redisDSN, err = redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}

	cleanup = func() {
		if err := dbContainer.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate postgres container: %v", err)
		}
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate redis container: %v", err)
		}
	}

	return pgDSN, redisDSN, cleanup
}

const testJWTSecret = "test-secret-key-minimum-32-bytes!!"

// newTestAPI wires up the app using the new Setup logic for integration tests
func newTestAPI(ctx context.Context, pgDSN string, redisDSN string) (http.Handler, func(), error) {
	_ = os.Setenv("DATABASE_DSN", pgDSN)
	_ = os.Setenv("REDIS_DSN", redisDSN)
	_ = os.Setenv("JWT_SECRET", testJWTSecret)
	_ = os.Setenv("LOG_LEVEL", "error")

	cfg := config.Load()
	logger := observability.NewLogger(cfg.Log.Level, cfg.Log.Format, "test")

	return app.Setup(ctx, cfg, logger)
}

// registerAndLogin calls the register endpoint and returns the access token.
func registerAndLogin(api http.Handler, email, password string) (string, error) {
	body := map[string]string{"email": email, "password": password}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", time.Now().UnixNano()%255)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		return "", fmt.Errorf("register failed: %s", w.Body.String())
	}

	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		return "", err
	}
	return resp.Data.AccessToken, nil
}

func doRequest(api http.Handler, method, path, token string, body []byte, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", time.Now().UnixNano()%255)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func buildMultipartBody(fields map[string]string) ([]byte, string) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	_ = mw.Close()
	return buf.Bytes(), mw.FormDataContentType()
}
