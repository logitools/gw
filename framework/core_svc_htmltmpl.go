package framework

import (
	"path/filepath"

	"github.com/logitools/gw/tpl"
)

// PrepareHTMLTemplateStore builds a new HTMLTemplateStore and atomically publishes it.
// It may be safely invoked while the app is running to hot-reload templates without restarting the server.
func (c *Core) PrepareHTMLTemplateStore() error {
	store := tpl.NewHTMLTemplateStore()

	if err := store.LoadFileTemplates(filepath.Join(c.AppRoot, "templates", "html")); err != nil {
		return err
	}

	// LATER
	// if err := store.BuildDerived(); err != nil { return err }

	// publish only after fully built = single pointer assignment as the last step
	c.HTMLTemplateStore = store
	return nil
}
