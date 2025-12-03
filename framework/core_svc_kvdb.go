package framework

import (
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"

	"github.com/logitools/gw/db/kvdb/impls/redis"
)

func (c *Core[B]) PrepareKVDatabase() error {
	// Load KV Database Config File
	err := c.loadKVDBConf()
	if err != nil {
		return err
	}
	if err = c.prepareKVDBClient(); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) loadKVDBConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".kv-databases.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.KVDBConf); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) prepareKVDBClient() error {
	switch c.KVDBConf.Type {
	case "redis":
		c.BackendKVDBClient = &redis.Client{Conf: &c.KVDBConf}
		if err := c.BackendKVDBClient.Init(); err != nil {
			return err
		}
	// case "memcached"
	default:
		return errors.New("unsupported key-value database type")
	}
	return nil
}
