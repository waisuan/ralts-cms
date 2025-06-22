.PHONY: generate fmt build run test clean start-dynamodb create-table setup-env

generate:
	go generate ./...

fmt:
	go fmt ./...

build:
	go build -o bin/ralts-cms cmd/web/main.go

run:
	go run cmd/web/main.go

test:
	go test ./...

clean:
	rm -rf bin/

start-dynamodb:
	./scripts/start-dynamodb.sh

create-table:
	./scripts/create-table.sh

setup-env:
	@if [ ! -f .env.development ]; then \
		echo "Creating .env.development from env.example..."; \
		cp env.example .env.development; \
		echo "Environment file created. Please review and modify .env.development as needed."; \
	else \
		echo ".env.development already exists."; \
	fi

setup: setup-env start-dynamodb create-table
	sleep 5
	echo "Setup complete! DynamoDB is running and table is created."

dev: setup run
