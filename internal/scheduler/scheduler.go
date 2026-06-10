package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/steemfans/go-pricefeed/internal/config"
	"github.com/steemfans/go-pricefeed/internal/log"
	"github.com/steemfans/go-pricefeed/internal/price"
	"github.com/steemfans/go-pricefeed/internal/publisher"
)

// Scheduler runs periodic price feed cycles.
type Scheduler struct {
	cfg        *config.Config
	aggregator *price.Aggregator
	publisher  *publisher.Publisher
	mu         sync.Mutex
	running    bool
	idle       sync.Cond
}

// New creates a scheduler.
func New(cfg *config.Config, aggregator *price.Aggregator, pub *publisher.Publisher) *Scheduler {
	s := &Scheduler{
		cfg:        cfg,
		aggregator: aggregator,
		publisher:  pub,
	}
	s.idle.L = &s.mu
	return s
}

// Run executes cycles until ctx is cancelled and the active cycle finishes.
func (s *Scheduler) Run(ctx context.Context) error {
	s.startCycle(ctx)

	ticker := time.NewTicker(time.Duration(s.cfg.Interval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.waitIdle()
			return ctx.Err()
		case <-ticker.C:
			if !s.tryStartCycle() {
				log.Println("previous cycle still running, skipping tick")
				continue
			}
			go s.startCycle(ctx)
		}
	}
}

func (s *Scheduler) tryStartCycle() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	return true
}

func (s *Scheduler) startCycle(ctx context.Context) {
	s.mu.Lock()
	if !s.running {
		s.running = true
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.idle.Broadcast()
		s.mu.Unlock()
	}()

	s.runCycle(ctx)
}

func (s *Scheduler) waitIdle() {
	s.mu.Lock()
	for s.running {
		s.idle.Wait()
	}
	s.mu.Unlock()
}

func (s *Scheduler) runCycle(ctx context.Context) {
	prices, err := s.aggregator.Collect(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		log.Println("no prices found.")
		return
	}

	avg := price.Average(prices)
	log.Printf("%v", prices)
	log.Printf("Price = %v", avg)

	if err := s.publisher.Publish(ctx, avg); err != nil && ctx.Err() == nil {
		log.Printf("publish stopped: %v", err)
	}
}
