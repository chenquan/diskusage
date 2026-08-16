package worker

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorker_Run(t *testing.T) {
	w := New(4)

	var running, maxRunning int64
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		w.Run(func() {
			defer wg.Done()

			cur := atomic.AddInt64(&running, 1)
			for {
				old := atomic.LoadInt64(&maxRunning)
				if cur <= old || atomic.CompareAndSwapInt64(&maxRunning, old, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&running, -1)
		})
	}
	wg.Wait()

	if got := atomic.LoadInt64(&maxRunning); got < 2 {
		t.Errorf("expected tasks to run concurrently (max running >= 2), got %d", got)
	}
}

func TestWorker_Close(t *testing.T) {
	w := New(4)
	w.Close()

	// after Close, Run must execute synchronously (inline)
	done := make(chan struct{})
	w.Run(func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run after Close did not execute inline")
	}
}

func TestWorker_ConcurrentRunWithClose(t *testing.T) {
	// Run racing with Close must not panic (regression: Close used to
	// close the channel, and a Run between the flag check and the send
	// would panic with "send on closed channel").
	w := New(2)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				w.Run(func() {})
			}
		}()
	}
	w.Close()
	wg.Wait()
}
