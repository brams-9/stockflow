# StockFlow

Tech Stack:
- Backend: Go, Echo, PostgreSQL, sqlc
- Frontend: React, TypeScript, Vite, TanStack Query, zod

## Running with Docker Compose

From the directory root:

```
docker compose up --build
```

This starts everything in order:

When the containers are up, open http://localhost:5173.

The compose command at startup will seeds 16 products and 3 orders. It is safe to start the stack again, the seed itself is idempotent.

Quick healt check once it is running:

```
curl http://localhost:8080/health
```

## Backend structure

The backend architecure i use follows `Uncle Bob's clean architecture` . Uncle Bob's use `usecase` for business logic layer but in here the use case layer is called `service` .

```
backend/
├── cmd/api/                      entry point, wires everything together
├── db/
│   ├── migrations/               golang-migrate SQL files
│   ├── queries/                  sqlc query files
│   └── seed/                     seed.sql
└── internal/
    ├── domain/                   entities and domain errors, no framework or json tags
    ├── service/                  business rules (products, stock, orders)
    ├── repository/               repository interfaces and the transaction runner
    ├── infrastructure/
    │   ├── config/               environment config
    │   └── postgres/             pgx implementations of the repositories, sqlc generated code
    └── delivery/http/
        ├── handler/              Echo handlers
        ├── dto/                  request and response shapes, kept apart from domain entities
        ├── error_handler.go      maps domain errors to the error envelope
        └── router.go
```

## Frontend structure

The frontend is organised by domain, with the usual pages, components, hooks and utils folders around it.

```
frontend/src/
├── main.tsx                      providers and routes
├── pages/                        ProductsPage, CreateOrderPage, OrderDetailPage
├── domains/
│   ├── product/                  api calls, query hooks, product modals
│   ├── order/                    api calls and query hooks
│   └── summary/                  summary card calculations
├── components/                   shared UI (Modal, ConfirmDialog, Badge, states)
├── hooks/                        useDebounce, useIdempotencyKey
├── lib/                          apiClient (the only place that calls fetch), queryClient
├── types/api.ts                  zod schemas and the types inferred from them
└── utils/                        formatting, validation, idempotency key helper
```

## Tests

Backend Go tests (the integration tests need the compose database and are skipped without `TEST_DATABASE_URL`):

```
docker run --rm --network stockflow_default -v "$PWD/backend":/app -w /app \
  -e TEST_DATABASE_URL='postgres://stockflow:stockflow@db:5432/stockflow?sslmode=disable' \
  golang:1.27 go test -race -count=1 ./...
```

## Known limitations

- There is no endpoint to list orders in the task pdf, so an order is opened by id (the "Open order" box) or right after it is created.
