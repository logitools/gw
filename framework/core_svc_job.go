package framework

import (
	"github.com/logitools/gw/schedjobs"
)

func (c *Core) PrepareJobScheduler() {
	c.JobScheduler = schedjobs.NewScheduler(c.RootCtx)
	c.AddService(c.JobScheduler)
}
