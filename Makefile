.PHONY: run test test-integration lint migrate-up migrate-down docker-up docker-down docker-clean format tidy vet

# ── Dev ───────────────────────────────────────────────────────────────────────
run:
	go run ./cmd/server

# ── Quality ───────────────────────────────────────────────────────────────────
test:
	go test $$(go list ./... | grep -v /tests) -race -cover -coverprofile=coverage.out

test-verbose:
	go test $$(go list ./... | grep -v /tests) -race -v -cover

test-integration:
	go test ./tests/... -v -timeout 120s

coverage:
	go tool cover -html=coverage.out

format:
	goimports -w .

lint:
	golangci-lint fmt ./... && golangci-lint run ./...

vet:
	go vet ./...

tidy:
	go mod tidy

# ── Docker ────────────────────────────────────────────────────────────────────
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-clean:
	docker compose down -v

docker-logs:
	docker compose logs -f api

# ── Build ─────────────────────────────────────────────────────────────────────
build-frontend:
	@echo "🎨 Building frontend..."
	@cd frontend && npm install && npm run build
	@echo "📦 Copying frontend build to embed directory..."
	@rm -rf internal/web/dist
	@cp -r frontend/dist internal/web/dist
	@touch internal/web/dist/.gitkeep
	@echo "✅ Frontend built and ready for embedding"

build: build-frontend
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o bin/server ./cmd/server

clean:
	rm -rf bin/ coverage.out internal/web/dist
	mkdir -p internal/web/dist
	touch internal/web/dist/.gitkeep
