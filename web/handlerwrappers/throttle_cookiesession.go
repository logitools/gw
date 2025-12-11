package handlerwrappers

import (
	"net/http"
	"time"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/web/cookiesession"
	"github.com/logitools/gw/web/responses"
)

type ThrottleCookieSession struct {
	AppProvider   framework.AppProviderFunc
	BucketGroupID string
}

func (m *ThrottleCookieSession) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Prerequisite _ SessionID
		sessionID, ok := cookiesession.SessionIDFromContext(ctx)
		if !ok {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "invalid session ID")
			return
		}
		// Check Throttle Bucket
		if !appCore.ThrottleBucketStore.Allow(m.BucketGroupID, sessionID, time.Now()) {
			responses.WriteSimpleErrorJSON(w, http.StatusTooManyRequests, "session rate limited")
			return
		}

		// Inner
		inner.ServeHTTP(w, r)

		// Post-action
	})
}
