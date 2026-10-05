#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

VOLUME_NAME="financial-management-tracking_postgres_data"

echo "==> Resetting local database volume..."

docker compose down --remove-orphans --volumes || true

docker compose rm -sf postgres backend frontend 2>/dev/null || true

if docker volume inspect "${VOLUME_NAME}" >/dev/null 2>&1; then
    echo "Volume ${VOLUME_NAME} is already removed by compose down."
else
    echo "Volume ${VOLUME_NAME} does not exist. Continuing..."
fi

echo "==> Starting services and applying migrations from scratch..."
docker compose up -d --build

echo "==> Waiting for backend health checks..."
MAX_RETRIES=30
RETRY_COUNT=0

until curl -s http://localhost:8080/api/v1/health | grep -q '"status":"healthy"'; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ "${RETRY_COUNT}" -ge "${MAX_RETRIES}" ]; then
        echo "Error: Backend failed to become healthy after reset."
        docker compose logs backend
        exit 1
    fi
    echo "Waiting for backend... (${RETRY_COUNT}/${MAX_RETRIES})"
    sleep 2
done

echo "==> Database reset successful!"
echo "    Frontend: http://localhost:3000"
echo "    Backend:  http://localhost:8080"
echo "    Health:   http://localhost:8080/api/v1/health"
