package cookiesession

import (
	"context"
	"strconv"

	"github.com/logitools/gw/authuser"
)

type idCtxKey struct{}

func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, idCtxKey{}, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(idCtxKey{})
	val, ok := ctxVal.(string)
	return val, ok
}

func SessionIDToInt64UIDCtxInjector(ctx context.Context, sessionID string, uidStr string) (context.Context, error) {
	ctx = ContextWithSessionID(ctx, sessionID)
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return authuser.WithUserID[int64](ctx, uid), nil
}
