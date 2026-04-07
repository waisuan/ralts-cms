// Package integration holds HTTP integration tests (PostgreSQL via testcontainers, S3 via LocalStack).
//
// They are behind the "integration" build tag so default unit runs stay fast:
//
//	go test ./...
//
// Run integration tests locally (Docker required):
//
//	go test -tags=integration -count=1 -timeout=15m ./internal/integration/...
package integration
