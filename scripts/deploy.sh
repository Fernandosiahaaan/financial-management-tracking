#!/usr/bin/env bash
set -euo pipefail

# FinTrack Deployment Script
ENV_TARGET="${1:-prod}"

if [ "$ENV_TARGET" = "dev" ] || [ "$ENV_TARGET" = "development" ]; then
    ENV_FILE=".env.development"
    echo "==> Deploying FinTrack in DEVELOPMENT mode (using $ENV_FILE)..."
else
    ENV_FILE=".env.production"
    echo "==> Deploying FinTrack in PRODUCTION mode (using $ENV_FILE)..."
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: Configuration file $ENV_FILE not found!"
    exit 1
fi

if grep -q "\[DEV-PROJECT-REF\]" "$ENV_FILE"; then
    echo "⚠️  PERINGATAN: File $ENV_FILE masih berisi placeholder [DEV-PROJECT-REF]!"
    echo "Silakan isi DATABASE_URL dengan connection string Supabase Dev Anda di file $ENV_FILE terlebih dahulu."
    exit 1
fi

# Synchronize to .env so docker compose and local apps use it consistently
cp "$ENV_FILE" .env
cp "$ENV_FILE" backend/.env

# 1. Check docker & docker compose
if ! command -v docker &> /dev/null; then
    echo "Error: docker is required but not installed."
    exit 1
fi

# 2. Build and restart containers using targeted env file
echo "==> Building and launching containers via Docker Compose..."
docker compose --env-file "$ENV_FILE" down --remove-orphans || true
docker compose --env-file "$ENV_FILE" build --pull
docker compose --env-file "$ENV_FILE" up -d

# 3. Wait for services to become healthy
echo "==> Waiting for backend and database health checks..."
MAX_RETRIES=30
RETRY_COUNT=0

until curl -s http://localhost:8080/api/v1/health | grep -q '"status":"healthy"'; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ "${RETRY_COUNT}" -ge "${MAX_RETRIES}" ]; then
        echo "Error: Backend failed to become healthy within timeout!"
        docker compose logs backend
        exit 1
    fi
    echo "Waiting for services... (${RETRY_COUNT}/${MAX_RETRIES})"
    sleep 2
done

echo "==> Deployment SUCCESSFUL!"
echo "    Frontend: http://localhost:3000"
echo "    Backend:  http://localhost:8080"
echo "    Health:   http://localhost:8080/api/v1/health"
