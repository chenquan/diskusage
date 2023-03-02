package worker

import (
	"testing"
	"time"
)

func TestWorker_Run(t *testing.T) {
	worker := New(10)
	worker.Run(func() {
		time.Sleep(time.Second * 2)
	})
	worker.Close()

	worker.Run(func() {
		time.Sleep(time.Second * 2)
	})

}
