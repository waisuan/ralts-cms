package handlers

import (
	"context"

	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
)

// resolveUpdatedByUsername returns the username of the currently authenticated user, for
// stamping onto records on create/update. It never errors: if there's no authenticated
// user (e.g. anonymous access) or the user lookup fails, it returns an empty string so
// callers can proceed without blocking the write.
func resolveUpdatedByUsername(ctx context.Context, d *deps.Dependencies) string {
	userCtx, err := appctx.GetUserFromContext(ctx)
	if err != nil || userCtx == nil {
		return ""
	}

	user, err := d.UsersRepository.GetByID(ctx, userCtx.UserID)
	if err != nil || user == nil {
		return ""
	}

	return user.Username
}
