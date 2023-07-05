package task

type Worker struct {
	limitPerWorker int
	mode           limitMode
}

func (w *Worker) Setlimit(k int, mode limitMode) {
	w.limitPerWorker = k
	w.mode = mode
}
