DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=activity_diary
SSL_MODE=disable

API_DIR=services/api
ANALYTICS_DIR=services/analytics
MIGRATIONS_DIR=$(API_DIR)/internal/migrations

POSTGRESQL_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(SSL_MODE)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make test-api"
	@echo "  make test-analytics"
	@echo "  make test"
	@echo "  make docker-up"
	@echo "  make docker-down"
	@echo "  make docker-logs"
	@echo "  make migrate-up"
	@echo "  make migrate-down"
	@echo "  make migrate-force VERSION=1"
	@echo "  make migrate-create NAME=add_new_table"

.PHONY: test-api
test-api:
	cd $(API_DIR) && go test ./...

.PHONY: test-analytics
test-analytics:
	cd $(ANALYTICS_DIR) && go test ./...

.PHONY: test
test: test-api test-analytics

.PHONY: docker-up
docker-up:
	docker compose up --build

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f

.PHONY: migrate-up
migrate-up:
	migrate -database "$(POSTGRESQL_URL)" -path "$(MIGRATIONS_DIR)" up

.PHONY: migrate-down
migrate-down:
	migrate -database "$(POSTGRESQL_URL)" -path "$(MIGRATIONS_DIR)" down 1

.PHONY: migrate-force
migrate-force:
	$(if $(VERSION),,$(error VERSION is required. Usage: make migrate-force VERSION=1))
	migrate -database "$(POSTGRESQL_URL)" -path "$(MIGRATIONS_DIR)" force $(VERSION)

.PHONY: migrate-create
migrate-create:
	$(if $(NAME),,$(error NAME is required. Usage: make migrate-create NAME=add_new_table))
	migrate create -ext sql -dir "$(MIGRATIONS_DIR)" -seq "$(NAME)"
