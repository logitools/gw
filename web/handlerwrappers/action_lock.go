package handlerwrappers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/namedlocks"
	"github.com/logitools/gw/web/responses"
)

// ActionLock to acquire the named locks for some actions.
// Each lock's name is "actionName:targetValue" meaning an action on a target.
// @field Actions map[string]string: { "actionName":"targetKey", ... }
// If targetKey = "AuthUID", then targetValue = Auth UserID (string) of the request.
// Otherwise, targetValue = r.PathValue(targetKey) i.e. the value for {targetKey} in the request url
//
// If failed to acquire any of the required locks, the request is blocked (locked out)
// If succeeded to acquire all the required locks, it proceeds to the inner layer keeping other attempts locked out
// When everything's done in the inner layer, releasing the acquired locks is the Post-action.
type ActionLock struct {
	AppProvider        framework.AppProviderFunc
	Actions            map[string]string
	AuthUIDStrProvider func(context.Context) (string, error) // Optional. Required if any of targetKey = "AuthID"
}

func (m *ActionLock) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pre-action
		ctx := r.Context()

		// Auth UserID String (Optional)
		var (
			authUIDStr         string
			authUIDProviderErr error
		)

		if m.AuthUIDStrProvider != nil {
			authUIDStr, authUIDProviderErr = m.AuthUIDStrProvider(ctx)
			if authUIDProviderErr != nil {
				responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, authUIDProviderErr.Error())
				return
			}
		}

		var (
			lockKeys []string
			target   string
		)
		for action, targetKey := range m.Actions {
			if targetKey == "AuthUID" {
				if authUIDStr == "" {
					responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "AuthUID Provider Required")
					return
				}
				target = authUIDStr
			} else {
				target = r.PathValue(targetKey)
			}
			lockKeys = append(lockKeys, fmt.Sprintf("%s:%s", action, target))
		}

		acquired, ok := namedlocks.AcquireLocks(appCore.ActionLocks, lockKeys)
		if !ok {
			// Fail-fast: resource is already locked
			if len(lockKeys) == 1 {
				responses.WriteSimpleErrorJSON(w, http.StatusConflict, fmt.Sprintf("action [%s] locked by another request", lockKeys[0]))
				return
			}
			lockedActionsStr := strings.Join(lockKeys, ", ")
			responses.WriteSimpleErrorJSON(w, http.StatusConflict, fmt.Sprintf("some of actions in [%s] locked by another request", lockedActionsStr))
			return
		}

		defer func() { // defer = Post-action
			namedlocks.ReleaseLocks(appCore.ActionLocks, acquired)
			if rcv := recover(); rcv != nil {
				log.Printf("[PANIC] user=%s method=%s path=%s locks=%v err=%v",
					authUIDStr,
					r.Method,
					r.URL.Path,
					acquired,
					rcv,
				)
				// re-panic if you want to propagate
				// panic(rcv)
			}
		}()

		// new context for the next handler
		ctx = namedlocks.ContextWithAcquiredLocks(ctx, acquired)

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action -> check out the defer func() above
	})
}
