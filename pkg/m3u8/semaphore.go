package m3u8

import (
	"context"
)

// semaphore implements a counting semaphore using channels
type semaphore struct {
	sem chan struct{}
}

// newSemaphore creates a semaphore with n permits
func newSemaphore(n int) *semaphore {
	return &semaphore{
		sem: make(chan struct{}, n),
	}
}

// acquire blocks until a permit is available or context is cancelled
func (s *semaphore) acquire(ctx context.Context) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// release returns a permit to the semaphore
func (s *semaphore) release() {
	<-s.sem
}
