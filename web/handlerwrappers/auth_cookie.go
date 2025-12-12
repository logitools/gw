package handlerwrappers

import (
	"fmt"
	"net/http"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/web/cookiesession"
	"github.com/logitools/gw/web/responses"
)

type AuthCookie struct {
	AppProvider framework.AppProviderFunc
}

// Wrap middleware func
// Extracts the Session Cookie from the request and Decrypt it to the Session ID, and Find it in the KVDB.
func (m *AuthCookie) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cookieSessionMgr := appCore.CookieSessionManager
		// If Logged-in, Session Cookie must be shipped in the request
		sessionCookie, err := r.Cookie(cookiesession.CookieName)
		if err != nil { // http.ErrNoCookie
			// Session Cookie Not Found = Non-login Hit to Auth-protected Endpoints
			// Redirect to Login page setting Intended URI Cookie
			// ToDo: flash msg "Login Required"
			cookiesession.SetCookie(w, r, 60)
			http.Redirect(w, r, cookieSessionMgr.Conf.LoginPath+"?endpoint=protected", http.StatusSeeOther)
			return
		}
		sessionIDBytes, err := cookieSessionMgr.Cipher.DecodeDecrypt(sessionCookie.Value)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, fmt.Sprintf("invalid session. %v", err))
			return
		}
		sessionID := string(sessionIDBytes)

		found, err := cookieSessionMgr.SessionExistsInKVDB(ctx, sessionID)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to check session. %v", err))
			return
		}
		if !found {
			// Session Expired. Redirect to Login page Clearing Session Cookie
			cookieSessionMgr.RemoveSessionCookie(w)
			cookiesession.SetCookie(w, r, 60)
			http.Redirect(w, r, cookieSessionMgr.Conf.LoginPath+"?session=expired", http.StatusSeeOther)
			return
		}

		// new context for the next handler
		ctx = cookiesession.ContextWithSessionID(ctx, sessionID)

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action

	})
}
