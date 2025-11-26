package pgsql

import "github.com/logitools/gw/db/sqldb"

func Register() {
	sqldb.RegisterFactory(DBType, NewClient)
}
