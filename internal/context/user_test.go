package context_test

import (
	"context"
	"testing"

	appctx "ralts-cms/internal/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithUser_GetUserFromContext(t *testing.T) {
	t.Parallel()

	u := &appctx.UserContext{EntityID: "42", UserID: 42, Role: "ADMIN"}
	ctx := appctx.WithUser(context.Background(), u)

	got, err := appctx.GetUserFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.EntityID, got.EntityID)
	assert.Equal(t, u.UserID, got.UserID)
	assert.Equal(t, u.Role, got.Role)
}

func TestGetUserFromContext_Missing(t *testing.T) {
	t.Parallel()

	_, err := appctx.GetUserFromContext(context.Background())
	assert.Error(t, err)
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Parallel()

	u := &appctx.UserContext{EntityID: "99", UserID: 99, Role: "user"}
	ctx := appctx.WithUser(context.Background(), u)

	ptr := appctx.GetUserIDFromContext(ctx)
	require.NotNil(t, ptr)
	assert.Equal(t, "99", *ptr)
}

func TestGetUserIDFromContext_Anonymous(t *testing.T) {
	t.Parallel()

	assert.Nil(t, appctx.GetUserIDFromContext(context.Background()))
}
