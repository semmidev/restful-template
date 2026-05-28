#!/usr/bin/env bash
# deploy.sh — Manual deploy helper script (without CI/CD)
# Useful for: first deploy, emergency rollback, debugging
#
# Usage:
#   ./deploy.sh                    # Deploy latest image
#   ./deploy.sh sha-abc123         # Deploy specific image tag
#   SERVER_IP=x.x.x.x ./deploy.sh # Override server IP inline

set -euo pipefail

# ─────────────────────────────────────────
# Config — adjust these or set as env vars
# ─────────────────────────────────────────
SERVER_IP="${SERVER_IP:-}"
SERVER_USER="${SERVER_USER:-deploy}"
APP_DIR="${APP_DIR:-/home/deploy/app}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_ed25519}"
DOCKER_IMAGE="${DOCKER_IMAGE:-sammidev/restful-template}"
IMAGE_TAG="${1:-latest}"
DOCKERHUB_USERNAME="${DOCKERHUB_USERNAME:-}"
DOCKERHUB_TOKEN="${DOCKERHUB_TOKEN:-}"

# ─────────────────────────────────────────
# Validation
# ─────────────────────────────────────────
if [ -z "$SERVER_IP" ]; then
  echo "❌ SERVER_IP is not set"
  echo "   Export first: export SERVER_IP=your.server.ip"
  echo "   Or: SERVER_IP=x.x.x.x ./deploy.sh"
  exit 1
fi

echo "🚀 Deploying to $SERVER_USER@$SERVER_IP"
echo "   Image: $DOCKER_IMAGE:$IMAGE_TAG"
echo ""

# ─────────────────────────────────────────
# Copy latest docker-compose.yml to server
# ─────────────────────────────────────────
echo "📁 Uploading docker-compose.yml..."
scp -i "$SSH_KEY" -o StrictHostKeyChecking=no \
  "$(dirname "$0")/../docker-compose.yml" \
  "$SERVER_USER@$SERVER_IP:$APP_DIR/"

# ─────────────────────────────────────────
# Deploy
# ─────────────────────────────────────────
echo "🐳 Deploying container..."
ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" <<ENDSSH
  set -e
  cd $APP_DIR

  export IMAGE_TAG=$IMAGE_TAG
  export DOCKER_IMAGE=$DOCKER_IMAGE

  # Login to Docker Hub (if credentials provided)
  if [ -n "$DOCKERHUB_TOKEN" ] && [ -n "$DOCKERHUB_USERNAME" ]; then
    echo "$DOCKERHUB_TOKEN" | docker login -u "$DOCKERHUB_USERNAME" --password-stdin
  fi

  # Pull latest image from Docker Hub
  docker compose pull app

  # Rolling restart — keeps db and redis running
  docker compose up -d --no-deps --remove-orphans app

  # Run database migrations (embedded in the binary via migrate-up subcommand)
  echo "🔄 Running migrations..."
  docker compose exec -T app /server migrate-up 2>/dev/null || \
    echo "⚠️  Migration skipped (migration subcommand not found — may auto-run on startup)"

  # Cleanup old images
  docker image prune -f

  echo "✅ Deploy complete!"
  docker compose ps
ENDSSH

echo ""
echo "✅ Done! Image: $IMAGE_TAG"
echo "   Health check: curl http://$SERVER_IP/api/v1/health"
