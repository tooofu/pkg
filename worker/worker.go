package worker

import (
	"errors"
	"log"
	"runtime"
	"sync"
)

var (
	// ErrFull errr full
	ErrFull = errors.New("worker chan full")
)

// Worker worker
type Worker struct {
	ch     chan func()
	worker int
	waiter sync.WaitGroup
}

// New worker
func New(worker, size int) *Worker {
	if worker <= 0 {
		worker = 1
	}

	w := &Worker{
		ch:     make(chan func(), size),
		worker: worker,
	}

	w.waiter.Add(worker)
	for i := 0; i < worker; i++ {
		go w.proc()
	}
	return w
}

func (w *Worker) proc() {
	defer w.waiter.Done()
	for {
		f := <-w.ch
		if f == nil {
			return
		}
		wrapRecover(f)()
	}
}

func wrapRecover(f func()) (res func()) {
	res = func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64*1024)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("panic in cache proc, err=%s stack=%s", r, buf)
			}
		}()
		f()
	}
	return
}

// Save save
func (w *Worker) Save(f func()) (err error) {
	if f == nil {
		return
	}

	select {
	case w.ch <- f:
	default:
		err = ErrFull
	}
	return
}

// Close close
func (w *Worker) Close() (err error) {
	for i := 0; i < w.worker; i++ {
		w.ch <- nil
	}
	w.waiter.Wait()
	return
}
