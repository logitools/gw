package framework

import (
	"encoding/json/v2"
	"os"
	"path/filepath"

	"github.com/logitools/gw/clients"
)

// PrepareClientApps prepares ClientApps
// building a new clients.ClientAppConf map and swaps the atomic pointer for the ClientApps
// So, this can be invoked to Hot-Reload the ClientApps
func (c *Core) PrepareClientApps() error {
	var (
		err           error
		newClientApps map[string]clients.ClientAppConf
	)
	if newClientApps, err = c.newClientAppsConfMapFromFile(); err != nil {
		return err
	}
	c.ClientApps.Store(&newClientApps) // atomic store
	return nil
}

func (c *Core) newClientAppsConfMapFromFile() (map[string]clients.ClientAppConf, error) {
	confFilePath := filepath.Join(c.AppRoot, "config", ".clients.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return nil, err
	}
	var confMap map[string]clients.ClientAppConf
	if err = json.Unmarshal(confBytes, &confMap); err != nil {
		return nil, err
	}
	return confMap, nil
}

// GetClientAppConf reads a clients.ClientAppConf
// Uses a single atomic cpu instruction
func (c *Core) GetClientAppConf(id string) (clients.ClientAppConf, bool) {
	confMapPtr := c.ClientApps.Load()
	if confMapPtr == nil {
		return clients.ClientAppConf{}, false
	}
	conf, ok := (*confMapPtr)[id]
	conf.ID = id
	return conf, ok
}
