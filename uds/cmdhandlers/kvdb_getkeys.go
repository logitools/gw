package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type KvdbGetKeysHandler struct {
	AppProvider framework.AppProviderFunc
}

func (h *KvdbGetKeysHandler) GroupName() string {
	return "kvdb"
}

func (h *KvdbGetKeysHandler) Command() string {
	return "kvdb-get-keys"
}

func (h *KvdbGetKeysHandler) Desc() string {
	return "Print all the keys in KV database"
}

func (h *KvdbGetKeysHandler) Usage() string {
	return h.Command()
}

func (h *KvdbGetKeysHandler) HandleCommand(_ []string, w io.Writer) error {
	appCore := h.AppProvider().AppCore()
	kvDBClient := appCore.KVDBClient
	ctx := appCore.RootCtx
	var cursor any = nil
	for {
		keys, nextCursor, err := kvDBClient.ScanKeys(ctx, cursor, 1000)
		if err != nil {
			return err
		}
		for _, key := range keys {
			_, _ = fmt.Fprintln(w, key)
		}
		if nextCursor == nil {
			break // done
		}
		cursor = nextCursor
	}
	return nil
}
