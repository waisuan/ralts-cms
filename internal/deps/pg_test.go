package deps

import (
	"context"
	"strings"
	"testing"
)

func TestNewPostgresClient_EmptyDatabaseURL(t *testing.T) {
	t.Parallel()

	_, err := NewPostgresClient(context.Background(), &Config{})
	if err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL in error, got: %v", err)
	}
}
