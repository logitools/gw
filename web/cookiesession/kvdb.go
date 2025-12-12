package cookiesession

import (
	"context"
	"errors"
)

// SessionIDToUIDStrKVDBSingleValue for the case KVDB SessionID:UidStr(SingleValue)
func SessionIDToUIDStrKVDBSingleValue(ctx context.Context, sessionMgr *Manager, sessionID string) (string, error) {
	key := sessionMgr.WebSessionIDToKVDBKey(sessionID)
	uidStr, ok, err := sessionMgr.BackendKVDBClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("session not found")
	}
	return uidStr, nil
}
