package framework

import (
	"context"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/logitools/gw/apis/mainbackend"
	"github.com/logitools/gw/clients"
	"github.com/logitools/gw/db/kvdb"
	"github.com/logitools/gw/db/sqldb"
	"github.com/logitools/gw/schedjobs"
	"github.com/logitools/gw/storages"
	"github.com/logitools/gw/svc"
	"github.com/logitools/gw/throttle"
	"github.com/logitools/gw/tpl"
	"github.com/logitools/gw/uds"
	"github.com/logitools/gw/web"
	"github.com/logitools/gw/web/session"
)

// Core - common config
type Core struct {
	AppName             string                                           `json:"app_name"`
	Listen              string                                           `json:"listen"`     // HTTP Application Listen IP:PORT Address
	Host                string                                           `json:"host"`       // HTTP Host. Can be used to generate public url endpoints
	DebugOpts           DebugOpts                                        `json:"debug_opts"` // Debug Options
	AppRoot             string                                           `json:"-"`          // Filled from compiled paths
	RootCtx             context.Context                                  `json:"-"`          // Global Context with RootCancel
	RootCancel          context.CancelFunc                               `json:"-"`          // CancelFunc for RootCtx
	UDSService          *uds.Service                                     `json:"-"`          // PrepareUDSService
	JobScheduler        *schedjobs.Scheduler                             `json:"-"`          // PrepareJobScheduler
	WebService          *web.Service                                     `json:"-"`          // PrepareWebService
	ThrottleBucketStore *throttle.BucketStore                            `json:"-"`          // PrepareThrottleBucketStore
	VolatileKV          *sync.Map                                        `json:"-"`          // map[string]string
	SessionLocks        *sync.Map                                        `json:"-"`          // map[string]*sync.Mutex for ServiceSessions and WebSessions
	ActionLocks         *sync.Map                                        `json:"-"`          // map[string]struct{}
	StorageConf         storages.Conf                                    `json:"-"`          // LoadStorageConf
	BackendHttpClient   *http.Client                                     `json:"-"`          // for requests to external apis
	KVDBConf            kvdb.Conf                                        `json:"-"`          // loadKVDBConf
	BackendKVDBClient   kvdb.Client                                      `json:"-"`          // prepareKVDBClient
	SQLDBConfs          map[string]*sqldb.Conf                           `json:"-"`          // loadSQLDBConfs
	BackendSQLDBClients map[string]sqldb.Client                          `json:"-"`          // prepareSQLDBClients
	ClientApps          atomic.Pointer[map[string]clients.ClientAppConf] `json:"-"`          // [Hot Reload] PrepareClientApps
	WebSessionManager   *session.Manager                                 `json:"-"`          // PrepareWebSessions
	MainBackendClient   *mainbackend.Client                              `json:"-"`          // PrepareMainBackendClient
	HTMLTemplateStore   *tpl.HTMLTemplateStore                           `json:"-"`          // PrepareHTMLTemplateStore

	services []svc.Service // Services to Manage
	done     chan error
}

func (c *Core) AddService(s svc.Service) {
	log.Printf("[INFO] adding service: %s", s.Name())
	c.services = append(c.services, s)
	log.Printf("[INFO] total services: %d", len(c.services))
}

func (c *Core) StartServices() error {
	c.done = make(chan error, len(c.services))
	for _, s := range c.services {
		err := s.Start()
		if err != nil {
			return err
		}
		go func(s svc.Service) {
			err := <-s.Done()
			c.done <- err
		}(s) // pass the loop var to the param. otherwise, they are captured inside goroutine lazily
	}
	return nil
}

func (c *Core) WaitServicesDone() error {
	for i := 0; i < len(c.services); i++ {
		if err := <-c.done; err != nil {
			return err
		}
	}
	return nil
}

func (c *Core) StopServices() {
	for _, s := range c.services {
		s.Stop()
	}
}
