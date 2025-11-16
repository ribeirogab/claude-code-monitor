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
	pauseCh  chan bool
	paused   bool
}

// New creates a new Scheduler instance
func New(interval time.Duration, task func() error) *Scheduler {
	return &Scheduler{
		interval: interval,
		task:     task,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		pauseCh:  make(chan bool, 1),
		paused:   false,
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
			if !s.paused {
				if err := s.task(); err != nil {
					log.Printf("Error executing task: %v", err)
				}
			}
		case pauseState := <-s.pauseCh:
			s.paused = pauseState
			if pauseState {
				log.Println("Scheduler paused")
			} else {
				log.Println("Scheduler resumed")
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

// Pause pauses the scheduler
func (s *Scheduler) Pause() {
	select {
	case s.pauseCh <- true:
	default:
	}
}

// Resume resumes the scheduler
func (s *Scheduler) Resume() {
	select {
	case s.pauseCh <- false:
	default:
	}
}

// IsPaused returns whether the scheduler is currently paused
func (s *Scheduler) IsPaused() bool {
	return s.paused
}
