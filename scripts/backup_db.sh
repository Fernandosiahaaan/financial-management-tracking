#!/usr/bin/env bash
set -euo pipefail

# FinTrack Database Backup Script
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/fintrack_backup_${TIMESTAMP}.sql.gz"
CONTAINER_NAME="${CONTAINER_NAME:-fintrack-postgres}"
DB_USER="${POSTGRES_USER:-fintrack}"
DB_NAME="${POSTGRES_DB:-fintrack}"

mkdir -p "${BACKUP_DIR}"

echo "==> Starting FinTrack Database Backup at ${TIMESTAMP}..."

if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    docker exec -t "${CONTAINER_NAME}" pg_dump -U "${DB_USER}" "${DB_NAME}" | gzip > "${BACKUP_FILE}"
else
    echo "Container ${CONTAINER_NAME} not running, falling back to local pg_dump..."
    pg_dump -U "${DB_USER}" -d "${DB_NAME}" | gzip > "${BACKUP_FILE}"
fi

BACKUP_SIZE=$(ls -lh "${BACKUP_FILE}" | awk '{print $5}')
echo "==> Backup completed successfully!"
echo "    File: ${BACKUP_FILE}"
echo "    Size: ${BACKUP_SIZE}"
