package handlerwrappers

import (
	"net/http"
	"time"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/web/requests"
	"github.com/logitools/gw/web/responses"
)

type ThrottleIP struct {
	AppProvider   func() framework.Application
	BucketGroupID string
}

func (m *ThrottleIP) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Requested IP
		ip := requests.GetClientIP(r)
		// Check Throttle Bucket
		if !appCore.ThrottleBucketStore.Allow(m.BucketGroupID, ip, time.Now()) {
			responses.WriteSimpleErrorJSON(w, http.StatusTooManyRequests, "access rate limited - ip "+ip)
			return
		}

		// Inner
		inner.ServeHTTP(w, r)

		// Post-action
	})
}
