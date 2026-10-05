#!/usr/bin/env bash
set -euo pipefail

# FinTrack Switch Environment Helper
MODE="${1:-}"

if [ "$MODE" != "dev" ] && [ "$MODE" != "prod" ]; then
    echo "Usage: ./scripts/switch_env.sh [dev|prod]"
    echo "Example:"
    echo "  ./scripts/switch_env.sh dev   # Switch local .env to development"
    echo "  ./scripts/switch_env.sh prod  # Switch local .env to production"
    exit 1
fi

if [ "$MODE" = "dev" ]; then
    SOURCE_FILE=".env.development"
    echo "==> Switching to DEVELOPMENT environment..."
else
    SOURCE_FILE=".env.production"
    echo "==> Switching to PRODUCTION environment..."
fi

if [ ! -f "$SOURCE_FILE" ]; then
    echo "Error: $SOURCE_FILE not found!"
    exit 1
fi

# Copy to root .env and backend/.env for local running
cp "$SOURCE_FILE" .env
cp "$SOURCE_FILE" backend/.env

echo "==> Active environment is now: $MODE"
echo "    Copied $SOURCE_FILE -> .env and backend/.env"
