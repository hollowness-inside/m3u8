package m3u8

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

// acquire blocks until a permit is available
func (s *semaphore) acquire() {
	s.sem <- struct{}{}
}

// release returns a permit to the semaphore
func (s *semaphore) release() {
	<-s.sem
}
