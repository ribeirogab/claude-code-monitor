package scheduler

import "time"

// Scheduler manages periodic task execution
type Scheduler struct {
	interval time.Duration
	task     func() error
	stopCh   chan struct{}
}

// New creates a new Scheduler instance
func New(interval time.Duration, task func() error) *Scheduler {
	return &Scheduler{
		interval: interval,
		task:     task,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	// TODO: Implement scheduler logic
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
}
