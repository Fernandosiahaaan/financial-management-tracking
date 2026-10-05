# Personal Finance Management System

A personal finance management application designed to help users understand:

* How much money comes in
* Where money is allocated
* How much money is spent
* Where money currently resides
* How money moves between accounts
* How much financial assets the user owns
* Whether total assets / net worth are increasing or decreasing
* Whether spending stays within the planned budget

The system is designed as a **multi-user application** from the beginning, with a web application as the initial client and a mobile application planned for the future.

---

## 1. Product Goal

The main goal of this application is to provide a clear picture of a user's financial condition during each financial cycle.

The application should answer four fundamental questions:

> **1. Uang saya ada di mana?**
> Accounts

> **2. Uang saya digunakan untuk apa?**
> Budget & Allocation

> **3. Uang saya bergerak ke mana?**
> Transactions

> **4. Apakah kondisi keuangan saya membaik?**
> Assets, Net Worth & Reports

---

# 2. Core Concepts

The system separates several financial concepts that are often mixed together in personal finance applications.

## 2.1 Account ≠ Allocation

An **Account** represents where money physically or virtually exists.

Examples:

* BCA
* Permata
* GoPay
* Cash
* Investment Account

An **Allocation** represents what money is intended for.

Examples:

* Operational
* Saving
* Investment
* Others

An allocation can exist across multiple accounts.

For example:

```text
Saving Allocation
├── BCA       Rp 5,000,000
├── GoPay     Rp   500,000
└── Investment Account
            Rp 10,000,000
```

Therefore, an account must not permanently represent a specific financial purpose.

---

## 2.2 Transaction ≠ Account Balance

Account balance should be explainable by financial events.

Conceptually:

```text
Current Balance
=
Opening Balance
+ Income
+ Transfer In
- Expense
- Transfer Out
± Adjustments
```

The system should avoid having an unexplained balance that cannot be traced back to transactions.

---

## 2.3 Transfer ≠ Expense

Moving money between accounts does not reduce total assets.

Example:

```text
BCA → GoPay
Rp 1,000,000
```

This is a **Transfer**, not an Expense.

Before:

```text
BCA       Rp 10,000,000
GoPay     Rp    500,000
-----------------------
Total     Rp 10,500,000
```

After:

```text
BCA       Rp  9,000,000
GoPay     Rp  1,500,000
-----------------------
Total     Rp 10,500,000
```

Total assets remain unchanged.

---

## 2.4 Investment Gain ≠ Income

Money added to an investment account and investment performance must be distinguishable.

Example:

```text
Investment Capital     Rp 10,000,000
Current Value           Rp 11,500,000
-------------------------------------
Investment Gain         Rp  1,500,000
```

The Rp 1,500,000 gain should not be treated as ordinary income from a salary or freelance transaction.

---

## 2.5 Receivable ≠ Expense

When the user lends money to another person, the money is converted from cash into a receivable.

Example:

```text
BCA                  -Rp 2,000,000
Receivable - Budi    +Rp 2,000,000
----------------------------------
Total Assets          unchanged
```

When Budi pays the money back:

```text
BCA                  +Rp 2,000,000
Receivable - Budi    -Rp 2,000,000
```

Again, total assets remain unchanged.

Therefore:

* Lending money is not an expense.
* Receiving repayment is not new income.

---

# 3. Financial Cycle

The application supports a **user-selectable financial cycle**.

The financial cycle is not necessarily aligned with the calendar month.

Users can define a cycle start day.

Example:

```text
Cycle Start Day = 25
```

Then:

```text
25 Aug → 24 Sep
25 Sep → 24 Oct
25 Oct → 24 Nov
```

The actual transaction date is always preserved.

---

## 3.1 Transaction Date

Every transaction has an actual financial date.

Example:

```text
Transaction Date:
2026-08-30
```

Even if the user enters the transaction on September 10, the transaction remains dated August 30.

The system should distinguish:

```text
transaction_date
created_at
```

Where:

* `transaction_date` = when the financial event happened
* `created_at` = when the record was entered into the application

This allows backdated transactions.

---

## 3.2 Financial Cycle vs Calendar Month

The system supports both perspectives.

### Calendar Month

Used for historical reporting.

Example:

```text
August 2026
01 Aug → 31 Aug
```

### Financial Cycle

Used for budgeting and personal financial planning.

Example:

```text
25 Aug → 24 Sep
```

A transaction can therefore belong to:

* August calendar month
* 25 Aug → 24 Sep financial cycle

without changing its original transaction date.

---

# 4. Example Financial Cycle

Assume the user has:

```text
Financial Cycle:
25 Aug → 24 Sep
```

Income:

```text
25 Aug   Salary          Rp 10,000,000
03 Sep   Passive Income   Rp  2,000,000
15 Sep   Freelance        Rp  1,000,000
---------------------------------------
Total Income             Rp 13,000,000
```

The cycle has:

```text
Available Income = Rp 13,000,000
```

The user may allocate the money:

```text
Operational    Rp 4,000,000
Saving         Rp 5,000,000
Investment     Rp 3,000,000
Others         Rp 1,000,000
--------------------------------
Total          Rp 13,000,000
```

Additional income during the cycle may:

1. Increase the available budget
2. Go directly to saving
3. Go directly to investment
4. Be carried over to the next cycle

The system should support these possibilities without changing the original transaction.

---

# 5. Budget & Actual Spending

Budget represents the planned amount.

Actual spending represents what really happened.

Example:

```text
Operational Budget       Rp 4,000,000
Actual Spending          Rp 4,300,000
-------------------------------------
Variance                -Rp   300,000
```

The system must not modify the budget simply because the user overspent.

---

## 5.1 Overspending Example

Suppose:

```text
Operational Budget = Rp 4,000,000
```

Actual expenses become:

```text
Food          Rp 1,500,000
Transport     Rp   800,000
Entertainment Rp 1,000,000
Others        Rp 1,000,000
--------------------------------
Total         Rp 4,300,000
```

The user then transfers:

```text
Saving BCA → Operational GoPay
Rp 300,000
```

The system should show:

```text
Operational Budget     Rp 4,000,000
Operational Spending   Rp 4,300,000
Overspending            Rp   300,000
```

The transfer does not change the actual expense.

This allows the system to identify:

> The user had to take money from savings to cover operational spending.

This can later become an important financial health metric.

---

# 6. Accounts

Accounts represent locations where money exists.

Supported account types:

```text
BANK
E_WALLET
CASH
INVESTMENT
OTHER
```

Examples:

```text
BCA
GoPay
Permata
Cash
Investment Account
```

Each account should have:

* Name
* Type
* Opening balance
* Current balance
* Status
* Created timestamp

Account balance should be derived from financial transactions whenever possible.

---

# 7. Transactions

Transactions represent financial events.

The main transaction types are:

```text
INCOME
EXPENSE
TRANSFER
```

Future transaction types may include:

```text
INVESTMENT_BUY
INVESTMENT_SELL
INVESTMENT_GAIN
RECEIVABLE_CREATE
RECEIVABLE_PAYMENT
ADJUSTMENT
```

The exact transaction model will be finalized during database/ERD design.

---

# 8. Categories

Expenses and income may be categorized.

Example expense categories:

```text
Food
Transport
Housing
Entertainment
Shopping
Utilities
Others
```

Example income categories:

```text
Salary
Freelance
Business
Passive Income
Bonus
Others
```

Categories should be configurable by each user rather than hard-coded globally.

---

# 9. Investment

Investment tracking is designed to be simple in the initial version.

Supported investment types may include:

```text
STOCK
MUTUAL_FUND
GOLD
CRYPTO
OTHER
```

The system should distinguish between:

```text
Capital Invested
Current Value
Gain / Loss
```

Example:

```text
Capital Invested     Rp 20,000,000
Current Value        Rp 22,500,000
----------------------------------
Unrealized Gain       Rp 2,500,000
```

Automated market price integration is not required for the initial version.

Manual valuation can be used first.

---

# 10. Receivables

Receivables are supported because lending money is relevant to the user's financial condition.

Example:

```text
Receivable:
Budi
Amount:
Rp 2,000,000
```

Possible statuses:

```text
ACTIVE
OVERDUE
PARTIALLY_PAID
PAID
WRITTEN_OFF
```

Partial payment should be supported.

Example:

```text
Original Receivable    Rp 2,000,000
Payment                Rp   500,000
Remaining              Rp 1,500,000
```

If a receivable is written off, the financial loss should affect assets/net worth while preserving the original historical transaction.

---

# 11. Debt / Liabilities

Debt is **not part of the initial MVP priority**.

However, the architecture should not prevent it from being added later.

Future model:

```text
Net Worth = Total Assets - Total Liabilities
```

Possible future entities:

```text
liabilities
liability_payments
```

For MVP, the focus is primarily:

```text
Assets
Income
Expenses
Budget
Allocation
Investments
Receivables
```

---

# 12. Assets & Net Worth

The application tracks financial assets such as:

```text
Bank
E-Wallet
Cash
Investment
Receivables
```

Total assets can conceptually be calculated as:

```text
Total Assets
=
Bank
+ E-Wallet
+ Cash
+ Investment
+ Receivables
```

If liabilities are introduced later:

```text
Net Worth
=
Total Assets
- Total Liabilities
```

---

# 13. Asset Growth

One of the important goals of the application is to determine whether the user's financial assets are increasing or decreasing.

Example:

```text
Previous Cycle Assets     Rp 50,000,000
Current Cycle Assets      Rp 55,000,000
-----------------------------------------
Growth                     Rp  5,000,000
```

The dashboard should make this easy to understand.

Possible indicators:

```text
Asset Growth
+Rp 5,000,000
+10%
```

The system should eventually distinguish asset growth caused by:

* New savings
* New investment capital
* Investment gain
* Income
* Other asset changes

---

# 14. Allocation Tracking

The system should be able to track how much money is allocated to financial purposes.

Example:

```text
Operational     40%
Saving          35%
Investment      20%
Others           5%
```

Allocation should not be permanently tied to an account.

For example:

```text
Saving
├── BCA
├── Cash
└── Investment Account
```

This allows the user to understand both:

1. Where money physically resides
2. What the money is intended for

---

# 15. Dashboard

The dashboard should answer the most important financial questions quickly.

A possible dashboard structure:

```text
Current Financial Cycle
25 Aug → 24 Sep

Income
Rp 13,000,000

Expense
Rp  4,300,000

Available
Rp  8,700,000

Operational Budget
Rp 4,000,000 / Rp 4,300,000

Total Assets
Rp 55,000,000

Asset Growth
+Rp 5,000,000
```

The UI should prioritize:

* Clarity
* Simplicity
* Readability
* Minimal cognitive load
* Important information first

The application should be usable by people of different ages and levels of financial knowledge.

---

# 16. Reports

The initial reporting system should support:

## Monthly Report

Calendar-based report.

```text
01 Sep → 30 Sep
```

## Financial Cycle Report

User-defined cycle.

```text
25 Aug → 24 Sep
```

## Cash Flow

Shows:

```text
Income
Expense
Transfer
Net Cash Flow
```

Transfers should not be treated as income or expense.

## Budget vs Actual

Example:

```text
Category       Budget       Actual       Variance

Food           1,500,000    1,600,000    -100,000
Transport        800,000      700,000    +100,000
Entertainment    500,000      650,000    -150,000
```

## Asset Growth

Shows asset changes across cycles/months.

## Net Worth

Shows:

```text
Assets
Liabilities
Net Worth
```

Liabilities can initially remain empty until the debt feature is introduced.

---

# 17. High-Level Architecture

The initial architecture uses a **modular monolith**.

```text
                    ┌───────────────┐
                    │   Web Client  │
                    └───────┬───────┘
                            │
                            │ REST API
                            ▼
                  ┌─────────────────────┐
                  │       Go API        │
                  │  Modular Monolith   │
                  └──────────┬──────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │   PostgreSQL   │
                    └────────────────┘
```

Future architecture:

```text
                         ┌─────────────┐
                         │     Web     │
                         └──────┬──────┘
                                │
                         ┌──────▼──────┐
                         │             │
                         │   Go API    │
                         │             │
                         └──────┬──────┘
                                │
                         ┌──────▼──────┐
                         │ PostgreSQL  │
                         └──────┬──────┘
                                │
                 ┌──────────────┴──────────────┐
                 │                             │
          ┌──────▼──────┐              ┌───────▼───────┐
          │    Mobile   │              │ Integrations  │
          │     App     │              │    Future     │
          └─────────────┘              └───────────────┘
```

The backend remains the central source of truth.

---

# 18. Backend Architecture

Backend technology:

```text
Language: Go
API: REST
Database: PostgreSQL
```

The application uses a modular monolith instead of microservices.

Recommended structure:

```text
internal/
├── auth/
├── user/
├── account/
├── category/
├── transaction/
├── budget/
├── investment/
├── receivable/
├── report/
└── dashboard/
```

Each module should follow a layered architecture:

```text
HTTP Handler
     │
     ▼
 Service / Use Case
     │
     ▼
 Repository
     │
     ▼
 PostgreSQL
```

Responsibilities:

### Handler

Responsible for:

* HTTP request
* Authentication context
* Request validation
* HTTP response

### Service

Responsible for:

* Business rules
* Transaction orchestration
* Financial calculations
* Domain validation

### Repository

Responsible for:

* Database queries
* Persistence
* Database transactions

---

# 19. Initial Domain Model

High-level domain structure:

```text
USER
 │
 ├── SETTINGS
 │
 ├── ACCOUNTS
 │
 ├── CATEGORIES
 │
 ├── FINANCIAL CYCLES
 │
 ├── TRANSACTIONS
 │
 ├── BUDGETS / ALLOCATIONS
 │
 ├── INVESTMENTS
 │
 ├── RECEIVABLES
 │
 └── ASSET SNAPSHOTS
```

Potential database tables:

```text
users
user_settings

accounts
categories

budgets
budget_allocations

transactions
transaction_categories

investment_assets
investment_holdings
investment_valuations

receivables
receivable_payments

asset_snapshots

audit_logs
```

This list is a starting point and will be refined during ERD/database design.

---

# 20. API Design

The API should be versioned from the beginning.

Base path:

```text
/api/v1
```

Potential endpoints:

```text
/api/v1/auth
/api/v1/accounts
/api/v1/accounts/:id

/api/v1/transactions
/api/v1/transactions/:id

/api/v1/categories

/api/v1/budgets
/api/v1/budgets/:cycle

/api/v1/investments
/api/v1/investments/holdings

/api/v1/receivables

/api/v1/reports/monthly
/api/v1/reports/cashflow
/api/v1/reports/networth

/api/v1/dashboard
```

The exact API contract will be defined after the domain model and database schema are finalized.

---

# 21. Multi-User Design

Although this is initially a personal finance application, the system is designed for multiple users.

Every user-owned financial entity must be scoped by:

```text
user_id
```

Example:

```text
User A
 ├── Account A
 ├── Transaction A
 └── Budget A

User B
 ├── Account B
 ├── Transaction B
 └── Budget B
```

User A must never be able to access User B's financial data by changing an entity ID in an API request.

Authorization must therefore be enforced at the service/repository level and not rely only on the frontend.

---

# 22. Security

Financial data is sensitive and must be treated accordingly.

Initial security requirements:

* Password hashing
* HTTPS
* Authentication
* Session/JWT-based authorization
* Refresh token strategy where applicable
* Input validation
* SQL injection protection
* User-level authorization
* Secure secret management
* Audit logging for important changes
* No plaintext passwords
* No sensitive information in application logs

Every financial query should be scoped to the authenticated user.

---

# 23. Money Representation

Money must **not** use floating-point types such as:

```text
float32
float64
```

for financial calculations.

Preferred approach for Rupiah:

```text
BIGINT
```

Example:

```text
Rp 10,500,000
```

stored as:

```text
10500000
```

This avoids floating-point precision problems.

The backend should have clear rules for:

* Money parsing
* Money formatting
* Addition/subtraction
* Validation
* Serialization in API responses

The exact representation will be finalized during database schema design.

---

# 24. Historical Data & Auditability

Financial records should be treated as historical facts.

The system should avoid casually deleting or modifying historical transactions.

For example:

```text
Transaction:
Rp 500,000
Food
2026-09-01
```

If the user entered it incorrectly, the system should preferably support correction while maintaining an auditable history.

Important records may use:

```text
created_at
updated_at
deleted_at
```

and/or an audit mechanism depending on the final design.

The goal is:

> A user's financial history should remain explainable.

---

# 25. Backdated Transactions

Backdated transactions are supported.

Example:

```text
Transaction entered:
10 Sep 2026

Actual transaction:
30 Aug 2026
```

The system stores:

```text
transaction_date = 2026-08-30
created_at       = 2026-09-10
```

The transaction therefore belongs to the financial cycle containing August 30.

This is important because users may forget to record transactions immediately.

---

# 26. Core Business Rules

The following rules are fundamental to the system.

### Rule 1 — Transfer does not change total assets

```text
Account A → Account B
```

changes account balances but not total assets.

### Rule 2 — Expense reduces assets

```text
Account → Expense
```

reduces total assets.

### Rule 3 — Income increases assets

```text
Income → Account
```

increases total assets.

### Rule 4 — Lending money does not immediately reduce total assets

```text
Cash → Receivable
```

changes the composition of assets.

### Rule 5 — Receivable repayment does not create new income

```text
Receivable → Cash
```

converts one asset type into another.

### Rule 6 — Investment gains/losses affect asset value

```text
Investment Value ↑
```

can increase total assets without being ordinary income.

### Rule 7 — Budget does not change because of overspending

```text
Budget = planned amount
Actual = real spending
```

Overspending is represented as variance.

### Rule 8 — Financial cycle does not change transaction date

The cycle is a reporting/planning boundary.

It does not replace the actual transaction date.

### Rule 9 — Account does not define allocation

A single account may contain money intended for multiple purposes.

### Rule 10 — User data is isolated

All financial data must belong to a user.

---

# 27. MVP Scope

## Phase 1 — MVP

### Authentication

* Register
* Login
* Logout
* Profile

### Accounts

* Create account
* Update account
* Archive account
* Opening balance
* Account balance

### Transactions

* Income
* Expense
* Transfer
* Transaction history
* Backdated transactions
* Categories

### Financial Cycle

* Configure cycle start day
* Current cycle
* Previous cycle
* Cycle summary

### Budget & Allocation

* Create budget
* Define categories
* Define allocations
* Budget vs actual

### Dashboard

* Income
* Expense
* Remaining budget
* Account balances
* Total assets
* Asset growth

### Receivables

* Create receivable
* Record payment
* Track remaining balance
* Receivable status

### Investment

* Investment account
* Holdings
* Cost basis
* Current valuation
* Gain/loss

---

# 28. Future Features

The following features are intentionally postponed:

```text
Recurring Transactions
CSV / Bank Statement Import
Automatic Bank Integration
Automatic Market Prices
Investment Automation
Notifications
Mobile Application
Advanced Debt / Liability Tracking
Financial Health Score
Advanced Analytics
```

Debt/liability tracking can later be added using the same architectural principles.

---

# 29. Suggested Development Order

Implementation should follow the dependency between domains.

```text
1. Project Setup
        ↓
2. Authentication
        ↓
3. User & Settings
        ↓
4. Accounts
        ↓
5. Categories
        ↓
6. Transactions
        ↓
7. Financial Cycle
        ↓
8. Budget & Allocation
        ↓
9. Dashboard
        ↓
10. Receivables
        ↓
11. Investments
        ↓
12. Reports
        ↓
13. Asset Snapshots
```

Database design should be finalized before implementing complex business logic.

---

# 30. Development Principles

### Keep the domain model explicit

Avoid hiding financial logic inside controllers.

Bad:

```text
HTTP Handler
    ↓
SQL Query
```

Preferred:

```text
Handler
   ↓
Service
   ↓
Repository
```

### Business rules belong to the domain/service layer

For example:

```text
Transfer
```

should be handled as one business operation that updates the source and destination account consistently.

### Use database transactions for financial operations

Operations that must succeed or fail together should use PostgreSQL transactions.

Example:

```text
Transfer BCA → GoPay

BEGIN

Create transfer
Decrease BCA
Increase GoPay

COMMIT
```

If any operation fails:

```text
ROLLBACK
```

No partial financial state should remain.

---

# 31. Design Philosophy

The application is intentionally designed around a simple financial mental model:

```text
                    INCOME
                      │
                      ▼
              AVAILABLE MONEY
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
     OPERATIONAL    SAVING    INVESTMENT
          │           │           │
          ▼           ▼           ▼
       EXPENSE     RESERVE      ASSET
                                  │
                                  ▼
                            GAIN / LOSS
```

While physical money is tracked separately:

```text
                    MONEY
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
       BANK         E-WALLET       CASH
                                     
                      +
                INVESTMENT
                      +
                 RECEIVABLE
```

This separation is one of the most important architectural decisions in the system.

---

# 32. Final Architecture Summary

```text
┌──────────────────────────────────────────────┐
│                  WEB CLIENT                  │
└──────────────────────┬───────────────────────┘
                       │
                       │ REST API
                       ▼
┌──────────────────────────────────────────────┐
│                    GO API                    │
│                                              │
│  Auth                                        │
│  User / Settings                             │
│  Accounts                                    │
│  Categories                                  │
│  Transactions                                │
│  Financial Cycle                             │
│  Budget / Allocation                         │
│  Investments                                 │
│  Receivables                                 │
│  Reports                                     │
│  Dashboard                                   │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                  PostgreSQL                  │
│                                              │
│  Users                                       │
│  Accounts                                    │
│  Transactions                                │
│  Budgets                                     │
│  Investments                                 │
│  Receivables                                 │
│  Snapshots                                   │
│  Audit Logs                                  │
└──────────────────────────────────────────────┘
```

Future clients:

```text
                    ┌── Web
                    │
PostgreSQL ← Go API ┼── Mobile
                    │
                    └── Integrations
```

The backend remains the single source of truth for all financial data.

---

# 33. Current Status

The project is currently in the **System Design / Domain Modeling phase**.

Next steps:

1. Finalize domain entities
2. Define entity relationships
3. Create ERD
4. Define database constraints
5. Finalize PostgreSQL schema
6. Define transaction model in detail
7. Define financial cycle calculation
8. Define budget/allocation rules
9. Define REST API contracts
10. Start Go project implementation

The database schema should **not** be implemented blindly from the initial table list above. The ERD and business rules should be finalized first.

---

## Core Principle

> **Know where the money is, know what it is for, know how it moves, and know whether your financial assets are growing.**
