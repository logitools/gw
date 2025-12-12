package handlerwrappers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/logitools/gw/contxt"
	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/web/cookiesession"
	"github.com/logitools/gw/web/responses"
)

type AuthCookieUser struct {
	AppProvider       framework.AppProviderFunc
	SessionIDToUIDStr func(context.Context, *cookiesession.Manager, string) (string, error)
	CtxInjector       contxt.BinaryInjectorFunc[string, string]
}

// Wrap middleware func
// Extracts the Session Cookie from the request and Decrypt it to the Session ID, and Find it in the KVDB.
func (m *AuthCookieUser) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cookieSessionMgr := appCore.WebSessionManager
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
		sessionIdBytes, err := cookieSessionMgr.Cipher.DecodeDecrypt(sessionCookie.Value)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, fmt.Sprintf("invalid session. %v", err))
			return
		}
		sessionID := string(sessionIdBytes)

		uidStr, err := m.SessionIDToUIDStr(ctx, cookieSessionMgr, sessionID)
		if err != nil {
			// Error or Not Found (maybe Session Expired)
			// Redirect to Login page Clearing Session Cookie
			cookieSessionMgr.RemoveWebSessionCookie(w)
			cookiesession.SetCookie(w, r, 60)
			http.Redirect(w, r, cookieSessionMgr.Conf.LoginPath+"?session=expired", http.StatusSeeOther)
			return
		}

		// new context for the next handler
		ctx, err = m.CtxInjector(ctx, sessionID, uidStr)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to inject data to the context. %v", err))
			return
		}

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action

	})
}
