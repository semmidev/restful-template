.PHONY: run test test-integration lint migrate-up migrate-down docker-up docker-down tidy vet

# ── Dev ───────────────────────────────────────────────────────────────────────
run:
	go run ./cmd/server

# ── Quality ───────────────────────────────────────────────────────────────────
test:
	go test ./... -race -cover -coverprofile=coverage.out

test-verbose:
	go test ./... -race -v -cover

test-integration:
	go test ./tests/... -v -timeout 120s

coverage:
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

vet:
	go vet ./...

tidy:
	go mod tidy

# ── Docker ────────────────────────────────────────────────────────────────────
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f api

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o bin/server ./cmd/server

clean:
	rm -rf bin/ coverage.out
