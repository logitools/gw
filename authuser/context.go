package authuser

import (
	"context"
)

type ctxKey struct{}

func WithUserID[T comparable](ctx context.Context, id T) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func UserIdFromContext[T comparable](ctx context.Context) (T, bool) {
	ctxVal := ctx.Value(ctxKey{})
	val, ok := ctxVal.(T)
	return val, ok
}
