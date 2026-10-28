package api

import (
	"context"

	"github.com/abudemy/backend/internal/service"
)

type ctxKey int

const (
	principalKey ctxKey = iota
	requestIDKey
)

// withPrincipal stores the authenticated principal in the request context.
func withPrincipal(ctx context.Context, p service.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// principalFrom extracts the authenticated principal from the context.
func principalFrom(ctx context.Context) (service.Principal, bool) {
	p, ok := ctx.Value(principalKey).(service.Principal)
	return p, ok
}

// mustPrincipal returns the principal or the zero value. Callers behind the
// auth middleware can safely assume it exists.
func mustPrincipal(ctx context.Context) service.Principal {
	p, _ := principalFrom(ctx)
	return p
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
