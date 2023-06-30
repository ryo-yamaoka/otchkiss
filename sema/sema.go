package sema

import (
	"context"

	"golang.org/x/sync/semaphore"
)

// Sema implements semaphore with it can be unlimited by specifying 0
type Sema struct {
	sem *semaphore.Weighted
}

// NewWeighted creates a new weighted semaphore with the given maximum combined weight for concurrent access.
// When you specify 0, it means unlimited concurrent access.
func NewWeighted(n int64) *Sema {
	var sem *semaphore.Weighted
	if n != 0 {
		sem = semaphore.NewWeighted(n)
	}
	return &Sema{
		sem: sem,
	}
}

// Acquire acquires the semaphore with a weight of n, blocking until resources are available or ctx is done. On success, returns nil. On failure, returns ctx.Err() and leaves the semaphore unchanged.
// If ctx is already done, Acquire may still succeed without blocking.
// If you created semaphore with 0, this method is non blocking.
func (s *Sema) Acquire(ctx context.Context, n int64) error {
	if s.sem == nil {
		return ctx.Err()
	}
	return s.sem.Acquire(ctx, n)
}

// Release releases the semaphore with a weight of n.
// But if you created semaphore with 0, this method is no effect.
func (s *Sema) Release(n int64) {
	if s.sem == nil {
		return
	}
	s.sem.Release(n)
}
