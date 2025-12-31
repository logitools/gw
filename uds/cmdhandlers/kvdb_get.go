package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type KvdbGetHandler struct {
	AppProvider framework.AppProviderFunc
}

func (h *KvdbGetHandler) GroupName() string {
	return "kvdb"
}

func (h *KvdbGetHandler) Command() string {
	return "kvdb-get"
}

func (h *KvdbGetHandler) Desc() string {
	return "Print the value of the given key in KV database"
}

func (h *KvdbGetHandler) Usage() string {
	return h.Command() + " key"
}

func (h *KvdbGetHandler) HandleCommand(args []string, w io.Writer) error {
	argLen := len(args)
	if argLen != 1 {
		return fmt.Errorf("usage: %s", h.Usage())
	}
	key := args[0]
	appCore := h.AppProvider().AppCore()
	kvDBClient := appCore.KVDBClient
	ctx := appCore.RootCtx
	typeName, err := kvDBClient.Type(ctx, key)
	if err != nil {
		return err
	}
	switch typeName {
	case "string":
		strVal, found, err := kvDBClient.Get(ctx, key)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("key not found")
		}
		_, _ = fmt.Fprintln(w, strVal)
	case "list":
		vals, err := kvDBClient.Range(ctx, key, 0, -1)
		if err != nil {
			return err
		}
		for _, v := range vals {
			_, _ = fmt.Fprintln(w, v)
		}
	case "hash":
		valMap, err := kvDBClient.GetAllFields(ctx, key)
		if err != nil {
			return err
		}
		for k, v := range valMap {
			_, _ = fmt.Fprintf(w, "%s: %s\n", k, v)
		}
	default:
		return fmt.Errorf("unsupported type: %s", typeName)
	}
	return nil
}
