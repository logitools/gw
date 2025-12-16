package framework

import (
	"path/filepath"

	"github.com/logitools/gw/tpl"
)

func (c *Core) PrepareHTMLTemplateStore() error {
	c.HTMLTemplateStore = tpl.NewHTMLTemplateStore()
	return c.HTMLTemplateStore.LoadFileTemplates(
		filepath.Join(c.AppRoot, "templates", "html"),
	)
}
