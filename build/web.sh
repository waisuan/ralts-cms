#!/bin/sh
set -eu

# golang-migrate CLI for pre-deploy; keep in sync with github.com/golang-migrate/migrate/v4 in go.mod
MIGRATE_VERSION="v4.19.0"
MIGRATE_URL="https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz"

curl -fsSL "$MIGRATE_URL" | tar -xz migrate
chmod +x migrate
go build -o bin/ralts-cms ./cmd/web
