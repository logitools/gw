package handlerwrappers

import (
	"context"
	"net/http"
	"time"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/web/responses"
)

type ThrottleUser struct {
	AppProvider    framework.AppProviderFunc
	UIDStrProvider func(context.Context) (string, error)
	BucketGroupID  string
}

// Wrap the middleware func
// prerequisite: UserID in the Request Context _ e.g. accesstoken.APIAccessTokenSession
func (m *ThrottleUser) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// UserID String
		uidStr, err := m.UIDStrProvider(ctx)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Check Throttle Bucket
		if !appCore.ThrottleBucketStore.Allow(m.BucketGroupID, uidStr, time.Now()) {
			responses.WriteSimpleErrorJSON(w, http.StatusTooManyRequests, "rate limited")
			return
		}

		// Inner
		inner.ServeHTTP(w, r)

		// Post-action
	})
}
