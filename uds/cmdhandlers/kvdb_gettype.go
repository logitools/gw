package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type KvdbGetType struct {
	AppProvider framework.AppProviderFunc
}

func (h *KvdbGetType) GroupName() string {
	return "kvdb"
}

func (h *KvdbGetType) Command() string {
	return "kvdb-get-type"
}

func (h *KvdbGetType) Desc() string {
	return "Print the type of the given key in KV database"
}

func (h *KvdbGetType) Usage() string {
	return h.Command() + " key"
}

func (h *KvdbGetType) HandleCommand(args []string, w io.Writer) error {
	argLen := len(args)
	if argLen != 1 {
		return fmt.Errorf("usage: %s", h.Usage())
	}
	key := args[0]

	appCore := h.AppProvider().AppCore()
	kvDBClient := appCore.KVDBClient
	ctx := appCore.RootCtx

	found, err := kvDBClient.Exists(ctx, key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("key not found")
	}
	typeName, err := kvDBClient.Type(ctx, key)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w, typeName)
	return nil
}
