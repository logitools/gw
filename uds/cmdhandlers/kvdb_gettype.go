package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type KvdbGetTypeHandler struct {
	AppProvider framework.AppProviderFunc
}

func (h *KvdbGetTypeHandler) GroupName() string {
	return "kvdb"
}

func (h *KvdbGetTypeHandler) Command() string {
	return "kvdb-get-type"
}

func (h *KvdbGetTypeHandler) Desc() string {
	return "Print the type of the given key in KV database"
}

func (h *KvdbGetTypeHandler) Usage() string {
	return h.Command() + " key"
}

func (h *KvdbGetTypeHandler) HandleCommand(args []string, w io.Writer) error {
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
