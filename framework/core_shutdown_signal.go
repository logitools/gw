package framework

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var once sync.Once

func (c *Core) startShutdownSignalListener() {
	once.Do(func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			log.Printf("[INFO] got signal [%s]. shutting down app [%s] ...", sig, c.AppName)
			c.RootCancel() // broadcast to all child services via Context.Done()
		}()
	})
	log.Printf("[INFO][CORE] shutdown signal listener started")
}
