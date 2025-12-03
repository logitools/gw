package framework

import (
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"

	"github.com/logitools/gw/apis/mainbackend"
)

// PrepareMainBackendClient to Send Request to the Main Backend API if any
// Prerequisite: BackendHttpClient
func (c *Core[B]) PrepareMainBackendClient() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".main-backend-api.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if c.BackendHttpClient == nil {
		return errors.New("backend http client not ready")
	}
	c.MainBackendClient = &mainbackend.Client{
		Client: c.BackendHttpClient,
	}
	if err = json.Unmarshal(confBytes, &c.MainBackendClient.Conf); err != nil {
		return err
	}
	return nil
}
