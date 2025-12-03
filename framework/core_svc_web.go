package framework

import (
	"net/http"

	"github.com/logitools/gw/web"
)

func (c *Core[B]) PrepareWebService(addr string, httpHandler http.Handler) {
	c.WebService = web.NewService(c.RootCtx, addr, httpHandler)
	c.AddService(c.WebService)
}
