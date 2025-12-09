package framework

import (
	"encoding/json/v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/logitools/gw/db/sqldb"
	"github.com/logitools/gw/db/sqldb/impls/mysql"
	"github.com/logitools/gw/db/sqldb/impls/pgsql"
)

func (c *Core) loadSQLDBConfs() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".sql-databases.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	c.SQLDBConfs = make(map[string]*sqldb.Conf)
	if err = json.Unmarshal(confBytes, &c.SQLDBConfs); err != nil {
		return err
	}
	return nil
}

// prepareSQLDBClients - Build & Init SQL DB Clients
// Use after loadSQLDBConfs
func (c *Core) prepareSQLDBClients() error {
	c.BackendSQLDBClients = make(map[string]sqldb.Client)

	// Registering Supported Implementations
	pgsql.Register()
	mysql.Register()

	// Prepare New Clients
	for dbName, sqlDBConf := range c.SQLDBConfs {
		log.Printf("[INFO][SQLDB] preparing %q", dbName)
		dbClient, err := sqldb.New(sqlDBConf.Type, sqlDBConf)
		if err != nil {
			return err
		}
		if err = dbClient.Init(); err != nil {
			return err
		}
		c.BackendSQLDBClients[dbName] = dbClient
	}
	return nil
}

// PrepareSQLDatabases for SQL DB Clients & RawSQL Stores, etc
func (c *Core) PrepareSQLDatabases(sqlFS fs.FS) error {
	// Load SQL Databases Config File
	err := c.loadSQLDBConfs()
	if err != nil {
		return err
	}
	DBTypesSet := make(map[string]struct{})
	for _, conf := range c.SQLDBConfs {
		DBTypesSet[conf.Type] = struct{}{}
	}
	if len(DBTypesSet) == 0 {
		return nil
	}

	// Prepare SQL DB Clients
	if err = c.prepareSQLDBClients(); err != nil {
		return err
	}

	if _, ok := DBTypesSet["mysql"]; ok {
		err = mysql.LoadRawStmtsToStore(sqlFS)
		if err != nil {
			return err
		}
	}
	if _, ok := DBTypesSet["pgsql"]; ok {
		err = pgsql.LoadRawStmtsToStore(sqlFS)
		if err != nil {
			return err
		}
	}
	return nil
}
