package cron

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

type executor struct {
	// wg *sync.WaitGroup // used by the scheduler to wait for the executor to stop
	log *logrus.Logger

	ctx    context.Context
	cancel context.CancelFunc

	jobFnCh chan Func
	jobWg   *sync.WaitGroup // used by the executor to wait for all jobs to finish
	// singletonWg *sync.Map

	limitPerConcurrency int

	stopped *atomic.Bool

	// locker Locker
}

func newExecutor(log *logrus.Logger) *executor {
	e := executor{
		// wg: &sync.WaitGroup{},
		log: log,

		jobFnCh: make(chan Func, 1000),
		jobWg:   &sync.WaitGroup{},

		stopped: atomic.NewBool(false),
	}
	e.ctx, e.cancel = context.WithCancel(context.Background())
	return &e
}

func execFn(f Func) {
	// TODO ignore counting for the moment
	// jf.runStartCount.Add(1)
	// fmt.Printf("%#v\n", jf)
	// fmt.Println(jf.isRunning)
	f.isRunning.Store(true)

	// callJobFunc(jf.eventListeners.onBeforeJobExecution)
	// _ = callJobFuncWithParams(jf.eventListeners.beforeJobRuns, []interface{}{jf.getName()})

	_ = callJobFuncWithParams(f.fn, f.params)
	// err := callJobFuncWithParams(jf.function, jf.params)
	// if err != nil {
	//     _ = callJobFuncWithParams(jf.eventListeners.onError, []interface{}{jf.getName(), err})
	// } else {
	//     _ = callJobFuncWithParams(jf.eventListeners.noError, []interface{}{jf.getName()})
	// }

	// _ = callJobFuncWithParams(jf.eventListeners.afterJobRuns, []interface{}{jf.getName()})
	// callJobFunc(jf.eventListeners.onAfterJobExecution)

	f.isRunning.Store(false)
	// jf.runFinishCount.Add(1)
}

func (e *executor) start() {
	// e.wg = &sync.WaitGroup{}

	for i := 0; i < e.limitPerConcurrency; i++ {
		e.jobWg.Add(1)

		// panicHandlerMu.RLock()
		// defer panicHandlerMu.RUnlock()
		// if panicHandler != nil {
		//     defer func() {
		//         if r := recover(); r != nil {
		//             panicHandler(f.funcName, r)
		//         }
		//     }()
		// }

		go e.run()
	}
}

func (e *executor) run() {
	for {
		select {
		case f := <-e.jobFnCh:
			e.runJob(f)
		case <-e.ctx.Done():
			e.jobWg.Done()
			return
		}
	}
}

func (e *executor) runJob(f Func) {
	// switch f.runConfig.mode {
	// case defaultMode:
	//     lockKey := f.jobName
	//     if lockKey == "" {
	//         lockKey = f.funcName
	//     }

	//     // if e.distributedLocker != nil {
	//     // }

	//     execJob(f)
	// case singletonMode:

	// }

	// e.log.Debugf("[executor.runJob] jobFnCh.len=%v\n", len(e.jobFnCh))
	execFn(f)
}

func (e *executor) stop() {
	e.stopped.Store(true)
	e.cancel()
	e.jobWg.Wait()
}
