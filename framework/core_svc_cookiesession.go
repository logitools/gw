package framework

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/logitools/gw/security"
	"github.com/logitools/gw/web/cookiesession"
)

// PrepareCookieSessions prepares CookieSessionManager
// Prerequisite: BackendKVDBClient
// Prerequisite: SessionLocks
func (c *Core) PrepareCookieSessions() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".web-session.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if c.BackendKVDBClient == nil {
		return errors.New("backend KVDB client not ready")
	}
	if c.SessionLocks == nil {
		return errors.New("sessionlocks not ready")
	}
	mgr := &cookiesession.Manager{
		AppName:           c.AppName,
		BackendKVDBClient: c.BackendKVDBClient,
		SessionLocks:      c.SessionLocks,
	}
	if err = json.Unmarshal(confBytes, &mgr.Conf); err != nil {
		return err
	}
	// Web Login Session Cipher
	cipher, err := security.NewXChaCha20Poly1305CipherBase64([]byte(mgr.Conf.EncryptionKey))
	if err != nil {
		return fmt.Errorf("NewXChaCha20Poly1305Cipher: %v", err)
	}
	mgr.Cipher = cipher

	c.CookieSessionManager = mgr
	return nil
}
