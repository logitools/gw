package handlerwrappers

import (
	"fmt"
	"net/http"

	"github.com/logitools/gw/clients"
	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/web/responses"
)

type APIClients struct {
	AppProvider framework.AppProviderFunc
	ClientIDs   map[string]struct{} // specific clients only. If nil, any clients
}

func (m *APIClients) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only pre-registered apps can use this api
		clientAppId := r.Header.Get("Client-Id") // string
		if clientAppId == "" {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "client app id required")
			return
		}
		if m.ClientIDs != nil {
			if _, ok := m.ClientIDs[clientAppId]; !ok {
				responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "client id blocked")
				return
			}
		}
		clientAppConf, ok := appCore.GetClientAppConf(clientAppId)
		if !ok {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, fmt.Sprintf("invalid client app: %s", clientAppId))
			return
		}
		ctx := clients.WithClientConf(r.Context(), clientAppConf)

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action
	})
}
