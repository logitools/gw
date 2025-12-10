package cookiesession

import (
	"context"
	"strconv"

	"github.com/logitools/gw/authuser"
)

type idCtxKey struct{}

func WithWebSessionId(ctx context.Context, webSessionID string) context.Context {
	return context.WithValue(ctx, idCtxKey{}, webSessionID)
}

func WebSessionIdFromContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(idCtxKey{})
	val, ok := ctxVal.(string)
	return val, ok
}

func SessionIDToInt64UIDCtxInjector(ctx context.Context, sessionID string, uidStr string) (context.Context, error) {
	ctx = WithWebSessionId(ctx, sessionID)
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return authuser.WithUserID[int64](ctx, uid), nil
}
