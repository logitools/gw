package framework

import (
	"github.com/logitools/gw/schedjobs"
)

func (c *Core[B]) PrepareJobScheduler() {
	c.JobScheduler = schedjobs.NewScheduler(c.RootCtx)
	c.AddService(c.JobScheduler)
}
