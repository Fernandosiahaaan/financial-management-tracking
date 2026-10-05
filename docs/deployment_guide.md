# Production Deployment Guide

This guide describes how to deploy, configure, secure, and maintain the **Personal Financial Management & Tracking System** in a production environment.

---

## 1. System Architecture

The application comprises three core components orchestrated via Docker Compose:

```
[ Internet / Clients ]
         │
         ▼
[ Reverse Proxy / Nginx (Port 80/443 SSL) ]
         │
    ┌────┴────────────────────────┐
    ▼                             ▼
[ Frontend (Nginx Static) ]   [ Backend API (Go) ]
Port: 3000                    Port: 8080
                                  │
                                  ▼
                      [ PostgreSQL 16 Database ]
                      Port: 5432 (Internal Docker Network)
```

- **Frontend**: Single Page Application built with Vite/React/TypeScript, served by Nginx with security headers, Gzip compression, and client-side routing.
- **Backend**: Go 1.24 REST API containerized with an unprivileged non-root user (`appuser:appgroup`), protected by security headers, rate limiting (300 req/min per IP), and max payload constraints.
- **Database**: PostgreSQL 16 Alpine, with persistent volume storage (`pgdata`), health checks, and connection pooling.

---

## 2. Prerequisites

- **Host Operating System**: Linux (Ubuntu 22.04 LTS / Debian 12 recommended) or macOS
- **Docker Engine**: v24.0+
- **Docker Compose**: v2.20+
- **Hardware Minimum**: 1 vCPU, 2GB RAM, 10GB SSD

---

## 3. Environment Configuration

Create a `.env` file in the project root based on the production configuration requirements:

```bash
# PostgreSQL Database Configuration
POSTGRES_DB=fintrack
POSTGRES_USER=fintrack_admin
POSTGRES_PASSWORD=GENERATE_A_STRONG_RANDOM_PASSWORD_HERE
DATABASE_URL=postgres://fintrack_admin:GENERATE_A_STRONG_RANDOM_PASSWORD_HERE@postgres:5432/fintrack?sslmode=disable

# Application Secrets
JWT_SECRET=GENERATE_A_SECURE_64_CHARACTER_RANDOM_KEY_HERE
PORT=8080

# Environment Mode
ENVIRONMENT=production
```

> **Security Note:** Never commit `.env` or production passwords to version control. Generate secrets using:
> ```bash
> openssl rand -base64 48
> ```

---

## 4. Production Deployment

### Quick Deploy via Script

A deployment script is provided at `scripts/deploy.sh`:

```bash
# Make executable (if not already)
chmod +x scripts/deploy.sh

# Run deployment
./scripts/deploy.sh
```

### Manual Deployment via Docker Compose

```bash
# 1. Pull / build the images
docker compose build --pull

# 2. Start services in detached mode
docker compose up -d

# 3. Verify container status and health
docker compose ps
```

---

## 5. Security & Hardening Checklist

1. **Non-Root Containers**:
   - Backend runs as UID/GID `10001` (`appuser`).
2. **Security Headers**:
   - Both Frontend and Backend enforce `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 1; mode=block`, `Referrer-Policy: strict-origin-when-cross-origin`, and `Content-Security-Policy`.
3. **Rate Limiting**:
   - Rate limiting is active on the Go API at 300 requests per minute per IP address with burst capability of 50.
   - Container healthchecks at `/api/v1/health` bypass rate limiting to prevent false alarms.
4. **Request Body Size Limit**:
   - All incoming requests are capped at 10MB to mitigate denial-of-service memory exhaustion.
5. **Network Isolation**:
   - PostgreSQL port `5432` is bound only to the internal Docker network or configured host port `5433` for administration. Ensure external firewall rules (UFW/iptables) block direct external access to Postgres.
6. **SSL/TLS Termination**:
   - In production, place a reverse proxy (e.g., Caddy, Traefik, or AWS ALB) in front of the application to terminate HTTPS with Let's Encrypt certificates.

---

## 6. Database Operations & Backups

### Automated Backups

Backups are executed using `scripts/backup_db.sh`:

```bash
./scripts/backup_db.sh
```

- Output: compressed dump file `backups/fintrack_backup_YYYYMMDD_HHMMSS.sql.gz`
- Automatic retention: archives older than 30 days are automatically pruned.

**Automate with Cron:**
```bash
# Run daily at 02:00 AM
0 2 * * * /opt/fintrack/scripts/backup_db.sh >> /var/log/fintrack_backup.log 2>&1
```

### Database Restoration

To restore from a backup archive:

```bash
./scripts/restore_db.sh backups/fintrack_backup_20261005_120000.sql.gz
```

The restore script prompts for confirmation before wiping and restoring the database.

---

## 7. Health Checks & Monitoring

- **Health Endpoint**:
  ```bash
  curl -i http://localhost:8080/api/v1/health
  ```
  Expected Response:
  ```json
  {"status":"ok","time":"2026-10-05T13:00:00Z"}
  ```
- **Container Health**:
  ```bash
  docker compose ps
  ```
- **Live Logs**:
  ```bash
  docker compose logs -f --tail=100 backend
  docker compose logs -f --tail=100 frontend
  ```

---

## 8. Rollback Procedure

If an update introduces unexpected regressions:

1. Restore previous application code / docker image tags:
   ```bash
   git checkout <previous-tag-or-commit>
   docker compose up --build -d
   ```
2. If schema changes need to be rolled back, restore the pre-deployment database backup:
   ```bash
   ./scripts/restore_db.sh backups/pre_deploy_backup.sql.gz
   ```
