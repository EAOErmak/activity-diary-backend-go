# Activity Diary Go

Reduced Go backend for the existing Activity Diary frontend.

## Services

- `services/api`: Gin REST API on `http://127.0.0.1:18080/api`
- `services/analytics`: internal analytics microservice on `http://127.0.0.1:18081`
- `postgres`: PostgreSQL 16

## Run

```bash
docker compose up --build
```

Seeded admin:

- `email`: `admin@example.com`
- `password`: `admin123`

## Tests

The project uses two Go modules:

```bash
cd services/api && go test ./...
cd ../analytics && go test ./...
```
