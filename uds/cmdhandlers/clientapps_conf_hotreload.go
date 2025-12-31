package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/logitools/gw/framework"
)

type ClientappsConfHotReload struct {
	AppProvider framework.AppProviderFunc
}

func (*ClientappsConfHotReload) GroupName() string {
	return "clientapps"
}

func (h *ClientappsConfHotReload) Command() string {
	return "clientapps-conf-hotreload"
}

func (h *ClientappsConfHotReload) Desc() string {
	return "Hot Reload Client Apps Conf"
}

func (h *ClientappsConfHotReload) Usage() string {
	return h.Command()
}

func (h *ClientappsConfHotReload) HandleCommand(_ []string, w io.Writer) error {
	appCore := h.AppProvider().AppCore()
	if err := appCore.PrepareClientApps(); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w, "client apps config hot-reloaded")
	return nil
}
