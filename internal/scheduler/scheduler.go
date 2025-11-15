package scheduler

import (
	"log"
	"time"
)

// Scheduler manages periodic task execution
type Scheduler struct {
	interval time.Duration
	task     func() error
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// New creates a new Scheduler instance
func New(interval time.Duration, task func() error) *Scheduler {
	return &Scheduler{
		interval: interval,
		task:     task,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	defer close(s.doneCh)

	// Execute immediately on start
	if err := s.task(); err != nil {
		log.Printf("Error executing task: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.task(); err != nil {
				log.Printf("Error executing task: %v", err)
			}
		case <-s.stopCh:
			log.Println("Scheduler stopped")
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
	<-s.doneCh
}
