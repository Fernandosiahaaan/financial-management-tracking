# Personal Financial Management & Tracking System

A robust, enterprise-grade personal financial management and tracking application built with a modern Go backend and a responsive React/TypeScript frontend.

---

## 🌟 Key Highlights & Domain Philosophy

This application strictly enforces financial domain invariants to ensure auditable accuracy and prevent common personal finance tracking pitfalls:

- **Account ≠ Allocation**: Physical account balances (bank accounts, e-wallets, cash) are distinct from virtual purpose allocations (needs, emergency, investments).
- **Transaction ≠ Account Balance**: Balances are calculated dynamically and reconciled against immutable transactions.
- **Transfer ≠ Expense**: Moving funds between accounts or allocations does not count as spending or revenue.
- **Investment Gain ≠ Income**: Unrealized and realized investment capital gains are tracked in a dedicated investment ledger, not blended into operational income.
- **Receivable ≠ Expense**: Money lent to others is an asset (receivable), not an operational expense.
- **Receivable Repayment ≠ Income**: Repayment of a receivable is an asset conversion, not income.
- **Zero Floating-Point Inaccuracies**: All monetary values are strictly modeled as 64-bit integers (`int64`) in Indonesian Rupiah (IDR).
- **Strict Multi-Tenant User Isolation**: Every query and database transaction is strictly scoped to the authenticated user ID.

---

## 🏗️ Architecture & Tech Stack

```
   ┌─────────────────────────────────────────────────────────┐
   │                   Browser / Client                      │
   └────────────────────────────┬────────────────────────────┘
                                │ HTTPS / REST (JSON)
                                ▼
   ┌─────────────────────────────────────────────────────────┐
   │               Frontend (React 19 + TypeScript)          │
   │  - Vite + TailwindCSS                                   │
   │  - Nginx 1.25 Alpine (Security Headers + Gzip)          │
   └────────────────────────────┬────────────────────────────┘
                                │ API Proxy (/api/v1/*)
                                ▼
   ┌─────────────────────────────────────────────────────────┐
   │               Backend API (Go 1.24)                     │
   │  - Clean Architecture: Handler → Service → Repository  │
   │  - Security: JWT Auth, Rate Limiter, Max Body Size      │
   │  - Non-root Container execution (UID 10001)             │
   └────────────────────────────┬────────────────────────────┘
                                │ Connection Pool (pgx)
                                ▼
   ┌─────────────────────────────────────────────────────────┐
   │             Database (PostgreSQL 16 Alpine)             │
   │  - Automated Migrations                                 │
   │  - ACID Transactions & Strict Foreign Keys              │
   └─────────────────────────────────────────────────────────┘
```

### Backend
- **Language**: Go 1.24
- **Router / Middleware**: Standard library `net/http` + token bucket rate limiter & security headers
- **Database Driver**: `github.com/lib/pq` / PostgreSQL 16
- **Authentication**: Stateless HMAC-SHA256 JWT tokens with `golang.org/x/crypto/bcrypt` password hashing
- **Security**: Content Security Policy, X-Frame-Options, XSS protection, 10MB payload limit, per-IP rate limiting

### Frontend
- **Framework**: React 19 with TypeScript
- **Tooling**: Vite 6
- **Styling**: Tailwind CSS & Modern Glassmorphism Design System
- **State & Data**: Custom API client with automatic token attachment and 401 handling
- **Testing**: Vitest & React Testing Library (76+ automated component tests)

---

## 🚀 Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) (v24.0+)
- [Docker Compose](https://docs.docker.com/compose/) (v2.20+)

### 1. Clone & Setup Environment

```bash
git clone https://github.com/username/financial-management-tracking.git
cd financial-management-tracking
cp .env.example .env # or verify docker-compose.yml defaults
```

### 2. Launch with Docker Compose

```bash
docker compose up --build -d
```

Services will be accessible at:
- **Frontend UI**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:8080/api/v1](http://localhost:8080/api/v1)
- **Health Check**: [http://localhost:8080/api/v1/health](http://localhost:8080/api/v1/health)

---

## 🧪 Testing

### Backend Unit & Integration Tests

```bash
cd backend
go test -v -race -cover ./...
```

All 12 backend packages include full test coverage (Accounts, Allocations, Auth, Budgets, Categories, Cycles, Dashboard, Health, Investments, Middleware, Receivables, Reports, Transactions).

### Frontend Component Tests

```bash
cd frontend
npm test -- --run
```

Runs all 23 Vitest test suites (76 unit/integration tests).

---

## 📖 API Documentation

Complete OpenAPI 3.0 specification is available at:
- [`docs/openapi.yaml`](docs/openapi.yaml)

To preview in Swagger Editor:
```bash
# Preview using any Swagger UI or OpenAPI extension
npx @redocly/cli preview-docs docs/openapi.yaml
```

---

## 🔒 Operations & Database Management

Automated shell scripts are provided in the `scripts/` directory:

- **Database Backup**:
  ```bash
  ./scripts/backup_db.sh
  ```
  Generates a timestamped `.sql.gz` dump under `backups/` and prunes backups older than 30 days.

- **Database Restore**:
  ```bash
  ./scripts/restore_db.sh backups/fintrack_backup_<timestamp>.sql.gz
  ```

- **Production Deployment**:
  ```bash
  ./scripts/deploy.sh
  ```

For detailed production instructions, see [`docs/deployment_guide.md`](docs/deployment_guide.md).

---

## 📂 Project Structure

```
├── .github/workflows/      # GitHub Actions CI workflow
├── backend/
│   ├── cmd/api/            # Application entrypoint & HTTP server
│   ├── internal/           # Domain modules (Clean Architecture)
│   │   ├── account/        # Accounts management
│   │   ├── allocation/     # Financial allocations & transfers
│   │   ├── auth/           # JWT authentication & user management
│   │   ├── budget/         # Budget caps and alerts
│   │   ├── category/       # Transaction categories
│   │   ├── cycle/          # Monthly / custom financial cycles
│   │   ├── dashboard/      # Unified financial overview metrics
│   │   ├── health/         # System liveness & readiness check
│   │   ├── investment/     # Investment assets & portfolio tracking
│   │   ├── middleware/     # Auth, Security Headers, Rate Limiter
│   │   ├── receivable/     # Money lent & debtor repayments
│   │   ├── report/         # Cashflow, cycle comparisons, exports
│   │   └── transaction/    # Operational income & expense entries
│   ├── migrations/         # PostgreSQL DDL migrations
│   └── Dockerfile          # Hardened Alpine build (non-root UID 10001)
├── frontend/
│   ├── src/
│   │   ├── components/     # UI modules & dashboards
│   │   ├── services/       # Typed API client
│   │   └── types/          # TypeScript domain models
│   ├── nginx.conf          # Hardened Nginx configuration
│   └── Dockerfile          # Multi-stage production build
├── docs/                   # System design, phases, and guides
│   ├── openapi.yaml        # OpenAPI 3.0 specification
│   ├── deployment_guide.md # Production deployment guide
│   └── system_design.md    # Architecture and domain models
├── scripts/                # Operations & maintenance scripts
└── docker-compose.yml      # Multi-service container orchestration
```

---

## 📄 License

This project is licensed under the MIT License.
