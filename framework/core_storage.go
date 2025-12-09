package framework

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
)

func (c *Core) LoadStorageConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".storages.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.StorageConf); err != nil {
		return err
	}
	return nil
}
