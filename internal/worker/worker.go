package worker

type Worker struct {
	worker chan struct{}
}

func New(capacity int) *Worker {
	return &Worker{worker: make(chan struct{}, capacity)}
}

func (w *Worker) Run(run func()) {
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
