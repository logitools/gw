package cookiesession

import (
	"context"
	"errors"
)

type KVDBBackendAPIData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserIDStr    string `json:"uid"`
	Key          string `json:"-"`
}

type KVDBUserIDData struct {
	UserIDStr string `json:"uid"`
	Key       string `json:"-"`
}

// SessionIDToUIDStrKVDBSingleValue for the case KVDB SessionID:UidStr(SingleValue)
func SessionIDToUIDStrKVDBSingleValue(ctx context.Context, sessionMgr *Manager, sessionID string) (string, error) {
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
