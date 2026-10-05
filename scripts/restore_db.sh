#!/usr/bin/env bash
set -euo pipefail

# FinTrack Database Restore Script
if [ $# -lt 1 ]; then
    echo "Usage: $0 <path-to-backup-file.sql.gz>"
    exit 1
fi

BACKUP_FILE="$1"
CONTAINER_NAME="${CONTAINER_NAME:-fintrack-postgres}"
DB_USER="${POSTGRES_USER:-fintrack}"
DB_NAME="${POSTGRES_DB:-fintrack}"

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "Error: Backup file '${BACKUP_FILE}' does not exist!"
    exit 1
fi

echo "==> Restoring FinTrack Database from: ${BACKUP_FILE}"
echo "    Target Database: ${DB_NAME} (User: ${DB_USER})"

if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    gunzip -c "${BACKUP_FILE}" | docker exec -i "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}"
else
    echo "Container ${CONTAINER_NAME} not running, falling back to local psql..."
    gunzip -c "${BACKUP_FILE}" | psql -U "${DB_USER}" -d "${DB_NAME}"
fi

echo "==> Database restore completed successfully!"
