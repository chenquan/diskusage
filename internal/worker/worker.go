//   Copyright 2023 chenquan
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package worker

import (
	"sync"
	"sync/atomic"
)

type Worker struct {
	worker    chan struct{}
	closeOnce sync.Once
	close     uint32
}

func New(capacity int) *Worker {
	return &Worker{worker: make(chan struct{}, capacity)}
}

func (w *Worker) Run(run func()) {
	if atomic.LoadUint32(&w.close) == 1 {
		run()
		return
	}

	select {
	case w.worker <- struct{}{}:
		go func() {
			defer func() {
				<-w.worker
			}()
			run()
		}()
	default:
		run()
	}
}

func (w *Worker) Close() {
	w.closeOnce.Do(func() {
		atomic.StoreUint32(&w.close, 1)
		close(w.worker)
	})
}
