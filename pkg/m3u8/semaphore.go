package m3u8

type Semaphore struct {
	sem chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, n),
	}
}

func (s *Semaphore) Acquire() {
	s.sem <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.sem
}
