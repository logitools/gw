package handlerwrappers

import (
	"fmt"
	"net/http"

	"github.com/logitools/gw/contxt"
	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/reason"
	"github.com/logitools/gw/security"
	"github.com/logitools/gw/web/responses"
)

type APIAccessToken struct {
	AppProvider       func() framework.Application
	UserIDCtxInjector contxt.InjectorFunc[string]
}

// Wrap is a middleware func
// Extracts the Access Token from the request header "Authorization", and Find it in the KVDB.
// If found, Injects the UID into the request context
func (m *APIAccessToken) Wrap(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pre-action
		appCore := m.AppProvider().AppCore()
		ctx := r.Context()
		// check the access access_token
		accessToken := security.ExtractBearerToken(r.Header.Get("Authorization")) // string
		if accessToken == "" {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "access token missing")
			return
		}
		key := appCore.AppName + "_access:" + security.HashHexSHA256(accessToken)
		uidStr, ok, err := appCore.BackendKVDBClient.Get(ctx, key)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch access token info. %v", err))
			return
		} else if !ok {
			responses.EncodeWriteJSON(w, http.StatusUnauthorized, responses.Message{
				Type:    "error",
				Message: "expired or invalid access token",
				Code:    reason.AccessTokenExpired,
			})
			return
		}

		ctx, err = m.UserIDCtxInjector(ctx, uidStr)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse uid. %v", err))
			return
		}

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action
	})
}
