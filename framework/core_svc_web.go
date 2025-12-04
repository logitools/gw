package framework

import (
	"net/http"

	"github.com/logitools/gw/web"
)

// PrepareWebService
// Call this when all the required services are prepared
func (c *Core[B]) PrepareWebService(addr string, httpHandler http.Handler) {
	c.WebService = web.NewService(c.RootCtx, addr, httpHandler)
	c.AddService(c.WebService)
}
