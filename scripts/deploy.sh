#!/usr/bin/env bash
set -euo pipefail

# FinTrack Deployment Script
echo "==> Deploying FinTrack Personal Finance Management System..."

# 1. Check docker & docker compose
if ! command -v docker &> /dev/null; then
    echo "Error: docker is required but not installed."
    exit 1
fi

# 2. Build and restart containers
echo "==> Building and launching containers via Docker Compose..."
docker compose down --remove-orphans || true
docker compose build --pull
docker compose up -d

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
