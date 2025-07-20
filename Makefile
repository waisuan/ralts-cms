.PHONY: generate fmt build run test clean setup-env \
	db-dev-up db-dev-down db-test-up db-test-down db-clean \
	test-with-db dev-with-db db-status db-logs

generate:
	go generate ./...

fmt:
	go fmt ./...

build:
	go build -o bin/ralts-cms cmd/web/main.go

run:
	APP_ENV=development go run cmd/web/main.go

test:
	APP_ENV=test go test ./...

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
