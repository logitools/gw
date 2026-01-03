package cookiesession

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
)

// GenerateSessionID generates 32 hex (0-9a-f) string from 16 random bytes for a Session ID
func GenerateSessionID() (string, error) {
	b := make([]byte, 16) // 128-bit random ID
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetCookie(w http.ResponseWriter, r *http.Request, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     IntendedURICookieName,
		Value:    url.QueryEscape(r.URL.RequestURI()),
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func RemoveCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     IntendedURICookieName,
		Path:     "/",
		MaxAge:   -1, // Delete
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// TryRedirectIfIntendedURICookie tries to redirect if IntendedURICookie is set returning true if redirected
// loginPath core.CookieSessionManager.Conf.LoginPath
func TryRedirectIfIntendedURICookie(w http.ResponseWriter, r *http.Request, loginPath string) bool {
	intendedUriCookie, err := r.Cookie(IntendedURICookieName)
	if err != nil {
		return false // no cookie [http.ErrNoCookie]
	}

	RemoveCookie(w)

	decodedURI, err := url.QueryUnescape(intendedUriCookie.Value)
	if err != nil || decodedURI == "" || decodedURI == loginPath {
		return false // malformed or meaningless value
	}

	if parsedURL, err := url.Parse(decodedURI); err != nil || parsedURL.IsAbs() {
		return false // reject external redirect
	}

	http.Redirect(w, r, decodedURI, http.StatusSeeOther)
	return true
}

func SessionIDToUIDStrFromKVDB(ctx context.Context, sessionMgr *Manager, sessionID string) (string, error) {
	key := sessionMgr.SessionIDToKVDBKey(sessionID)
	uidStr, ok, err := sessionMgr.KVDBClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("session not found")
	}
	return uidStr, nil
}
