package framework

import (
	"time"

	"github.com/logitools/gw/throttle"
)

func (c *Core[B]) PrepareThrottleBucketStore(cleanupCycle time.Duration, cleanupOlderThan time.Duration) {
	c.ThrottleBucketStore = throttle.NewBucketStore[B](c.RootCtx, cleanupCycle, cleanupOlderThan)
	c.AddService(c.ThrottleBucketStore)
}
