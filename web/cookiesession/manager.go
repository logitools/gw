package cookiesession

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/logitools/gw/db/kvdb"
	"github.com/logitools/gw/security"
)

type Manager struct {
	Conf         Conf
	Cipher       *security.XChaCha20Poly1305Cipher
	AppName      string // for session key, etc.
	SessionLocks *sync.Map
	KVDBClient   kvdb.Client
}

func (m *Manager) SessionIDToKVDBKey(sessionID string) string {
	return m.AppName + "_wsession:" + sessionID
}

func (m *Manager) SessionExistsInKVDB(ctx context.Context, sessionID string) (bool, error) {
	return m.KVDBClient.Exists(ctx, m.SessionIDToKVDBKey(sessionID))
}

func (m *Manager) CheckCookieSession(ctx context.Context, r *http.Request) bool {
	sessionCookie, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	cookieSessionId, err := m.Cipher.DecodeDecrypt(sessionCookie.Value) // []byte
	if err != nil {
		return false
	}
	found, err := m.SessionExistsInKVDB(ctx, string(cookieSessionId))
	if err != nil {
		return false
	}
	return found
}

func (m *Manager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	encSessionID, err := m.Cipher.EncryptEncode([]byte(sessionID))
	if err != nil {
		return fmt.Errorf("failed to encrypt web login session id. %v", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:  CookieName,
		Value: encSessionID,
		Path:  "/", // Subpaths will get this cookie.
		// Domain: // Cannot be set with `__Host-`
		HttpOnly: true, // JS cannot read it
		Secure:   true, // only sent over HTTPS
		MaxAge:   m.Conf.ExpireHardcap,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) RemoveSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		MaxAge:   -1, // Delete
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// StoreSessionInKVDBForBackendAPI stores a session in KVDB for Backend API and returns the session ID
// Session Key in KVDB stores "uid" for UserID and other backend API tokens.
func (m *Manager) StoreSessionInKVDBForBackendAPI(ctx context.Context, accessToken string, refreshToken string, uidStr string) (string, error) {
	cookieSessionID, err := GenerateSessionID()
	if err != nil {
		return "", err
	}
	// Store session_id in KvDB with access_token and refresh_token
	slidingExpiration := time.Duration(m.Conf.ExpireSliding) * time.Second
	hardcapExpiration := time.Duration(m.Conf.ExpireHardcap) * time.Second
	key := m.SessionIDToKVDBKey(cookieSessionID)
	if err = m.KVDBClient.SetFields(ctx, key, map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"uid":           uidStr,
	}); err != nil {
		return "", err
	}
	ok, err := m.KVDBClient.Expire(ctx, key, slidingExpiration)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("failed to set session expiration")
	}

	if m.Conf.MaxCntPerUser > 0 {
		usrSessionListKey := fmt.Sprintf("%s_wsessions:%s", m.AppName, uidStr)
		// SessionList Lock (User Level Lock)
		mu, _ := m.SessionLocks.LoadOrStore(usrSessionListKey, &sync.Mutex{})
		mutex := mu.(*sync.Mutex)

		mutex.Lock() // waits until this gets the lock if it's locked by another goroutine
		defer mutex.Unlock()
		// No need to delete the lock. Keep it for reusing it's tiny in memory.
		// By keeping it, no overhead to create(LoadOrStore)/delete the lock
		// If you wants to delete the lock every time:
		//defer func() {
		//	mutex.Unlock()
		//	env.SessionLocks.Delete(usrSessionListKey)
		//}()

		if err = m.KVDBClient.Push(ctx, usrSessionListKey, cookieSessionID); err != nil {
			return "", err
		}
		sessionCnt, err := m.KVDBClient.Len(ctx, usrSessionListKey)
		if err != nil {
			return "", err
		}
		if sessionCnt > m.Conf.MaxCntPerUser {
			diff := sessionCnt - m.Conf.MaxCntPerUser
			sessionsToDel, err := m.KVDBClient.Range(ctx, usrSessionListKey, 0, diff-1) // []string
			if err != nil {
				return "", err
			}
			var keysToDel []string
			for _, v := range sessionsToDel {
				keysToDel = append(keysToDel, m.SessionIDToKVDBKey(v))
			}
			_, _ = m.KVDBClient.Delete(ctx, keysToDel...)
			if err = m.KVDBClient.Trim(ctx, usrSessionListKey, diff, -1); err != nil {
				return "", err
			}
			_, _ = m.KVDBClient.Expire(ctx, usrSessionListKey, hardcapExpiration)
		}
	}

	return cookieSessionID, nil
}

func (m *Manager) KVDBBackendAPIData(ctx context.Context, key string) (*KVDBBackendAPIData, error) {
	sessionData, err := m.KVDBClient.GetAllFields(ctx, key)
	if err != nil {
		return nil, err
	}
	accessToken, ok1 := sessionData["access_token"]
	refreshToken, ok2 := sessionData["refresh_token"]
	uidStr, ok3 := sessionData["uid"]
	if !ok1 || !ok2 || !ok3 {
		return nil, errors.New("invalid session data")
	}
	return &KVDBBackendAPIData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserIDStr:    uidStr,
	}, nil
}

// StoreSessionInKVDBAsSingleValueUID stores a session with UID only (single value) in KVDB
// Returns the SessionID
// This is the Simplest Case -> No API Tokens in the KVDB for the Session
func (m *Manager) StoreSessionInKVDBAsSingleValueUID(ctx context.Context, uidStr string) (string, error) {
	cookieSessionID, err := GenerateSessionID()
	if err != nil {
		return "", err
	}
	// Store session_id in KvDB with access_token and refresh_token
	slidingExpiration := time.Duration(m.Conf.ExpireSliding) * time.Second
	hardcapExpiration := time.Duration(m.Conf.ExpireHardcap) * time.Second
	key := m.SessionIDToKVDBKey(cookieSessionID)
	if err = m.KVDBClient.Set(ctx, key, uidStr, slidingExpiration); err != nil {
		return "", err
	}

	if m.Conf.MaxCntPerUser > 0 {
		usrSessionListKey := fmt.Sprintf("%s_wsessions:%s", m.AppName, uidStr)
		// SessionList Lock (User Level Lock)
		mu, _ := m.SessionLocks.LoadOrStore(usrSessionListKey, &sync.Mutex{})
		mutex := mu.(*sync.Mutex)

		mutex.Lock() // waits until this gets the lock if it's locked by another goroutine
		defer mutex.Unlock()
		// No need to delete the lock. Keep it for reusing it's tiny in memory.
		// By keeping it, no overhead to create(LoadOrStore)/delete the lock
		// If you wants to delete the lock every time:
		//defer func() {
		//	mutex.Unlock()
		//	env.SessionLocks.Delete(usrSessionListKey)
		//}()

		if err = m.KVDBClient.Push(ctx, usrSessionListKey, cookieSessionID); err != nil {
			return "", err
		}
		sessionCnt, err := m.KVDBClient.Len(ctx, usrSessionListKey)
		if err != nil {
			return "", err
		}
		if sessionCnt > m.Conf.MaxCntPerUser {
			diff := sessionCnt - m.Conf.MaxCntPerUser
			sessionsToDel, err := m.KVDBClient.Range(ctx, usrSessionListKey, 0, diff-1) // []string
			if err != nil {
				return "", err
			}
			var keysToDel []string
			for _, v := range sessionsToDel {
				keysToDel = append(keysToDel, m.SessionIDToKVDBKey(v))
			}
			_, _ = m.KVDBClient.Delete(ctx, keysToDel...)
			if err = m.KVDBClient.Trim(ctx, usrSessionListKey, diff, -1); err != nil {
				return "", err
			}
			_, _ = m.KVDBClient.Expire(ctx, usrSessionListKey, hardcapExpiration)
		}
	}

	return cookieSessionID, nil
}

// SessionIDToUIDStrKVDBSingleValue for the case KVDB SessionID:UidStr(SingleValue)
// ToDo: make this as a method of Manager
func (m *Manager) SessionIDToUIDStrKVDBSingleValue(ctx context.Context, sessionID string) (string, error) {
	key := m.SessionIDToKVDBKey(sessionID)
	uidStr, ok, err := m.KVDBClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("session not found")
	}
	return uidStr, nil
}
