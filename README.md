# Orbit Employee Management System

A monorepo containing a Go API backend and a Next.js React frontend for managing users, appointments, payments, usage analytics, and Stripe-backed payment workflows.

## Architecture

The project is split into two main directories:
- `backend/`: A Go/Gin API handling business logic, Postgres database connections, Stripe webhooks, and Redis analytics.
- `frontend/`: A Next.js (App Router) web application using Tailwind CSS and TanStack Query to interface with the backend.

<!-- ```mermaid
flowchart LR
    Client[Next.js Frontend] --> Gin[Gin API]
    Gin --> Auth[JWT Middleware]
    Auth --> AppHandlers[Handlers]
    AppHandlers --> AppRepo[Postgres Repositories]
    AppHandlers --> Redis[(Redis Analytics)]
    AppHandlers --> Stripe[Stripe API]
    Stripe --> Webhook[Stripe Webhook Handler]
    Webhook --> PaymentRepo[Payments Repository]
    Cron[robfig/cron hourly job] --> AppRepo
    Cron --> EventRepo[Events Repository]
    AppRepo --> Postgres[(Postgres)]
    PaymentRepo --> Postgres
    EventRepo --> Postgres
    Redis --> Summary[Analytics Summary]
``` -->

## Prerequisites

- Go 1.25+
- Node.js 20+
- Docker and Docker Compose
- Stripe account with test mode enabled
- Stripe CLI for webhook testing

## Setup & Running Locally

### 1. Infrastructure (Database & Cache)

Start Postgres and Redis from the root directory:

```bash
docker-compose up -d
```

### 2. Backend API Setup

Navigate to the backend directory and set up your environment:

```bash
cd backend
cp .env.example .env
```

Run the database migrations in order:

```bash
psql "$POSTGRES_DSN" -f migrations/0001_init.up.sql
psql "$POSTGRES_DSN" -f migrations/0002_stripe_payments_webhooks.up.sql
```

Run the API:

```bash
go run .
```

### 3. Frontend Setup

In a new terminal, navigate to the frontend directory:

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:3000`.

## Webhook Testing

Use the Stripe CLI to forward test-mode webhook events to the local backend API:

```bash
stripe login
stripe listen --forward-to localhost:8080/api/webhooks/stripe
```

The CLI prints a webhook signing secret. Set that value as `STRIPE_WEBHOOK_SECRET` in `backend/.env`.

Then trigger a test payment flow or send a test event, for example:

```bash
stripe trigger payment_intent.succeeded
```
