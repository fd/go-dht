package kademlia

import (
	"log"
	"time"

	"golang.org/x/net/context"
)

var debug = false

func tokenStream(n int, duration time.Duration) <-chan struct{} {
	var out = make(chan struct{}, n)
	if n == 0 {
		close(out)
		return out
	}

	go func() {
		defer close(out)

		ticker := time.NewTicker(duration / time.Duration(n))
		defer ticker.Stop()

		for ; n > 0; n-- {
			<-ticker.C
			out <- struct{}{}
		}
	}()

	return out
}

type semaphore chan struct{}

func newSema(n int) semaphore {
	c := make(semaphore, n)
	for i := 0; i < n; i++ {
		c <- struct{}{}
	}
	return c
}

func (s semaphore) acquire(ctx context.Context) bool {
	select {
	case <-s:
		return true
	case <-ctx.Done():
		return false
	}
}

func (s semaphore) release() {
	s <- struct{}{}
}

func logf(format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
}
