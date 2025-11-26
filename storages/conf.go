package storages

import "github.com/logitools/gw/storages/keystores"

type Conf struct {
	KeyStoreConf keystores.Conf `json:"key_store"`
}
