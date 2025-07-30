package handlers

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
)

// MustFromContext retrieves the datastore from the context or panics if not found.
// This function is used with gin middleware to extract the datastore that was
// injected by the datastore middleware.
func MustFromContext(ctx context.Context) *pg.Datastore {
	// this is for gin middleware which does not accept key as any. only string.
	if c := ctx.Value("datastore"); c != nil {
		return c.(*pg.Datastore)
	}
	panic("datastore middleware did not inject datastore")
}
