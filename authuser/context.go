package authuser

import (
	"context"
	"errors"
	"strconv"
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

func StrUIDCtxInjector(ctx context.Context, uidStr string) (context.Context, error) {
	return WithUserID[string](ctx, uidStr), nil
}

func StrToInt64UIDCtxInjector(ctx context.Context, uidStr string) (context.Context, error) {
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return WithUserID[int64](ctx, uid), nil
}

func UidStrFromCtxInt64UID(ctx context.Context) (string, error) {
	UserID, ok := UserIdFromContext[int64](ctx)
	if !ok {
		return "", errors.New("no UserID in Context")
	}
	return strconv.FormatInt(UserID, 10), nil
}
