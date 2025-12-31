package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type SessionlockGetKeys struct {
	AppProvider framework.AppProviderFunc
}

func (h *SessionlockGetKeys) GroupName() string {
	return "sessionlock"
}

func (h *SessionlockGetKeys) Command() string {
	return "sessionlock-get-keys"
}

func (h *SessionlockGetKeys) Desc() string {
	return "Print session lock keys"
}

func (h *SessionlockGetKeys) Usage() string {
	return h.Command()
}

func (h *SessionlockGetKeys) HandleCommand(_ []string, w io.Writer) error {
	appCore := h.AppProvider().AppCore()
	sessionLocks := appCore.SessionLocks
	if sessionLocks == nil {
		return fmt.Errorf("session locks not ready")
	}
	sessionLocks.Range(func(k, _ any) bool {
		_, _ = fmt.Fprintln(w, k.(string))
		return true
	})
	return nil
}
