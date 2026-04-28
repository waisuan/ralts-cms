.PHONY: generate fmt build run test integration-test ui-test clean setup-env dev \
	db-dev-up db-dev-down db-test-up db-test-down db-clean \
	test-with-db dev-with-db db-status db-logs \
	migrate-dev migrate-test migrate-up migrate-down lint \
	server client ui-dev legacy-redirect-dev

generate:
	go generate ./...

fmt:
	go fmt ./...

lint:
	@if [ ! -f bin/revive ]; then \
		curl -L https://github.com/mgechev/revive/releases/latest/download/revive_$(shell uname -s | tr '[:upper:]' '[:lower:]')_$(shell uname -m | sed 's/x86_64/amd64/').tar.gz | tar -xz -C bin/ revive; \
		chmod +x bin/revive; \
	fi
	bin/revive -formatter friendly ./cmd/... ./internal/... ./pkg/...

build:
	go build -o bin/ralts-cms cmd/web/main.go

run:
	APP_ENV=development go run cmd/web/main.go

server:
	APP_ENV=development go run cmd/web/main.go

client:
	cd ui/ && npm run dev

# Install UI deps if missing, then start Next dev server
ui-dev:
	@if [ ! -d ui/node_modules ]; then \
		echo "Installing UI dependencies..."; \
		cd ui && npm install; \
	fi
	cd ui && npm run dev

# Legacy domain redirect (redirect-legacy/). Default PORT=3099; use localhost in the browser (not 127.0.0.1 on WSL/Windows).
legacy-redirect-dev:
	@if [ ! -d redirect-legacy/node_modules ]; then \
		echo "Installing redirect-legacy dependencies..."; \
		cd redirect-legacy && npm install; \
	fi
	@echo ""
	@echo "→ Open in browser: http://localhost:$(or $(PORT),3099)/"
	@echo ""
	cd redirect-legacy && PORT=$(or $(PORT),3099) npm start

test:
	APP_ENV=test go test ./...

# Jest unit tests in ui/ (installs ui/node_modules if missing)
ui-test:
	@if [ ! -d ui/node_modules ]; then \
		echo "Installing UI dependencies..."; \
		cd ui && npm install; \
	fi
	cd ui && npm test

# Full HTTP stack against Postgres + LocalStack (requires Docker). Not tagged on default `make test`.
integration-test:
	APP_ENV=test go test -tags=integration -count=1 -timeout=15m ./internal/integration/...

clean:
	rm -rf bin/

setup-env:
	@if [ ! -f .env.development ]; then \
		echo "Creating .env.development from env.example..."; \
		cp env.example .env.development; \
		echo "Environment file created. Please review and modify .env.development as needed."; \
	else \
		echo ".env.development already exists."; \
	fi
	@if [ ! -f .env.test ]; then \
		echo "Creating .env.test from env.example..."; \
		cp env.example .env.test; \
		echo "Environment file created. Please review and modify .env.test as needed."; \
	else \
		echo ".env.test already exists."; \
	fi

# Development setup
dev: setup-env db-up localstack-up
	@echo "Development environment setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Run migrations: make migrate-up"
	@echo "2. Start development server: make run"
	@echo "3. Or start with database: make dev-with-db"

localstack-up:
	docker compose -f docker-compose.dev.yml up -d localstack

# Database Management Commands

# Development Database
db-dev-up:
	@echo "Starting development PostgreSQL database..."
	docker compose -f docker-compose.dev.yml up -d postgres-dev
	@echo "Development database started on port 5432"
	@echo "Database: ralts_cms_dev"
	@echo "User: ralts_user"
	@echo "Password: ralts_password"

db-dev-down:
	@echo "Stopping development PostgreSQL database..."
	docker compose -f docker-compose.dev.yml down
	@echo "Development database stopped"

db-dev-logs:
	@echo "Showing development database logs..."
	docker compose -f docker-compose.dev.yml logs -f postgres-dev

# Test Database
db-test-up:
	@echo "Starting test PostgreSQL database..."
	docker compose -f docker-compose.test.yml up -d postgres-test
	@echo "Test database started on port 5433"
	@echo "Database: ralts_cms_test"
	@echo "User: ralts_user"
	@echo "Password: ralts_password"

db-test-down:
	@echo "Stopping test PostgreSQL database..."
	docker compose -f docker-compose.test.yml down
	@echo "Test database stopped"

db-test-logs:
	@echo "Showing test database logs..."
	docker compose -f docker-compose.test.yml logs -f postgres-test

# Combined Commands
db-up: db-dev-up db-test-up
	@echo "Both development and test databases are running"

db-down: db-dev-down db-test-down
	@echo "Both development and test databases stopped"

# Database Status
db-status:
	@echo "=== Development Database ==="
	@docker compose -f docker-compose.dev.yml ps
	@echo ""
	@echo "=== Test Database ==="
	@docker compose -f docker-compose.test.yml ps

# Clean up all database containers and volumes
db-clean:
	@echo "Cleaning up all database containers and volumes..."
	docker compose -f docker-compose.dev.yml down -v
	docker compose -f docker-compose.test.yml down -v
	@echo "All database containers and volumes removed"

# Development with database
dev-with-db: db-dev-up
	@echo "Starting development server with database..."
	@echo "Waiting for database to be ready..."
	@until docker compose -f docker-compose.dev.yml exec -T postgres-dev pg_isready -U ralts_user -d ralts_cms_dev; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "Database is ready!"
	$(MAKE) run

# Test with database
test-with-db: db-test-up
	@echo "Starting tests with test database..."
	@echo "Waiting for test database to be ready..."
	@until docker compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U ralts_user -d ralts_cms_test; do \
		echo "Waiting for test database..."; \
		sleep 2; \
	done
	@echo "Test database is ready!"
	@echo "Running tests..."
	POSTGRES_HOST=localhost POSTGRES_PORT=5433 POSTGRES_DB=ralts_cms_test $(MAKE) test
	@echo "Tests completed. Stopping test database..."
	$(MAKE) db-test-down

# Database connection helpers
db-dev-connect:
	@echo "Connecting to development database..."
	docker compose -f docker-compose.dev.yml exec postgres-dev psql -U ralts_user -d ralts_cms_dev

db-test-connect:
	@echo "Connecting to test database..."
	docker compose -f docker-compose.test.yml exec postgres-test psql -U ralts_user -d ralts_cms_test

# Migration Commands

# Run migrations on development database
migrate-dev:
	@echo "Running migrations on development database..."
	@echo "Waiting for development database to be ready..."
	@until docker compose -f docker-compose.dev.yml exec -T postgres-dev pg_isready -U ralts_user -d ralts_cms_dev; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "Database is ready! Running migrations..."
	migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms_dev?sslmode=disable" up
	@echo "Development database migrations completed!"

# Run migrations on test database
migrate-test:
	@echo "Running migrations on test database..."
	@echo "Waiting for test database to be ready..."
	@until docker compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U ralts_user -d ralts_cms_test; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "Database is ready! Running migrations..."
	migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5433/ralts_cms_test?sslmode=disable" up
	@echo "Test database migrations completed!"

# Run migrations on both databases
migrate-up: migrate-dev migrate-test
	@echo "All database migrations completed!"

# Rollback migrations on both databases
migrate-down:
	@echo "Rolling back migrations on development database..."
	migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5432/ralts_cms_dev?sslmode=disable" down
	@echo "Rolling back migrations on test database..."
	migrate -path ./db/migrations -database "postgresql://ralts_user:ralts_password@localhost:5433/ralts_cms_test?sslmode=disable" down
	@echo "All database migrations rolled back!"

seed-dev:
	APP_ENV=development go run cmd/cli/main.go -type machine -count 100
	APP_ENV=development go run cmd/cli/main.go -type user -count 10

# Append more machines to existing dev data without wiping.
# Usage: make seed-dev-append COUNT=50
seed-dev-append:
	APP_ENV=development go run cmd/cli/main.go -type machine -count $(or $(COUNT),50) -append