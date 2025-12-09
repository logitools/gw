package throttle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/logitools/gw/svc"
)

type BucketStore struct {
	Ctx              context.Context    // Service Context
	cancel           context.CancelFunc // Service Context CancelFunc
	state            int                // internal service state
	done             chan error         // Shutdown Error Channel
	cleanupCycle     time.Duration
	cleanupOlderThan time.Duration
	groups           map[string]*BucketGroup
}

func (s *BucketStore) Name() string {
	return "ThrottleBucketStore"
}

func NewBucketStore(parentCtx context.Context, cleanupCycle time.Duration, cleanupOlderThan time.Duration) *BucketStore {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &BucketStore{
		Ctx:              svcCtx,
		cancel:           svcCancel,
		state:            svc.StateREADY,
		done:             make(chan error, 1),
		cleanupCycle:     cleanupCycle,
		cleanupOlderThan: cleanupOlderThan,
		groups:           make(map[string]*BucketGroup),
	}
}

// Start starts a service that manages buckets
func (s *BucketStore) Start() error {
	if s.state == svc.StateRUNNING {
		return fmt.Errorf("already started")
	}
	if s.state != svc.StateREADY {
		return fmt.Errorf("cannot start. not ready")
	}
	s.state = svc.StateRUNNING
	log.Printf("[INFO][Throttle] cleanup service started cycle=%v exp=%v", s.cleanupCycle, s.cleanupOlderThan)
	go s.run()
	return nil
}

func (s *BucketStore) Stop() {
	if s.state != svc.StateRUNNING {
		log.Println("[ERROR][Throttle] cannot stop. not running")
		return
	}
	s.cancel()
	s.state = svc.StateSTOPPED
	log.Println("[INFO][Throttle] service stopped")
}

func (s *BucketStore) Done() <-chan error {
	return s.done
}

func (s *BucketStore) run() {
	ticker := time.NewTicker(s.cleanupCycle)
	defer ticker.Stop()
	for {
		select {
		case <-s.Ctx.Done():
			log.Println("[INFO][Throttle] stopping cleaning service")
			s.done <- nil
			return
		case now := <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PANIC] recovered in throttle bucketstore cleaning service: %v", r)
					}
				}()
				log.Printf("[INFO][Throttle] %v cleanup cycle ...", s.cleanupCycle)
				s.Cleanup(now)
			}()
		}
	}
}

func (s *BucketStore) GetBucketGroup(id string) (*BucketGroup, bool) {
	g, ok := s.groups[id]
	return g, ok
}

func (s *BucketStore) GetBucket(groupID string, bucketID string) (*Bucket, bool) {
	g, ok := s.groups[groupID]
	if !ok {
		return nil, false
	}
	return g.GetBucket(bucketID)
}

func (s *BucketStore) SetBucketGroup(id string, conf *BucketConf) {
	s.groups[id] = &BucketGroup{
		conf:    conf,
		buckets: &sync.Map{},
	}
}

func (s *BucketStore) Allow(groupID string, bucketID string, now time.Time) bool {
	g, ok := s.GetBucketGroup(groupID)
	if !ok {
		return false // Invalid groupID always Blocked
	}
	b, ok := g.GetBucket(bucketID)
	if ok {
		return b.Allow(now)
	}
	// consume 1 token from the fresh bucket
	g.SetBucket(bucketID, g.conf.Burst-1, now)
	return true
}

// Inspect returns a snapshot of all BucketGroup IDs and their local Bucket IDs.
// It does not lock globally, so results may be slightly inconsistent
// if buckets are being modified concurrently — which is fine for inspection.
func (s *BucketStore) Inspect() map[string][]string {
	result := make(map[string][]string)

	for groupID, bucketGroup := range s.groups {
		var ids []string
		bucketGroup.buckets.Range(func(localID, _ any) bool {
			ids = append(ids, localID.(string))
			return true
		})
		result[groupID] = ids
	}

	return result
}
