package cron

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.uber.org/atomic"
)

// // mode is the Job's running mode
// type mode int8

// const (
//     // defaultMode disable any mode
//     defaultMode mode = iota
//     // singletonMode switch to single job mode
//     singletonMode
// )

var (
	ErrTimeFormat           = errors.New("time format error")
	ErrParamsNotAdapted     = errors.New("the number of params is not adapted")
	ErrNotAFunction         = errors.New("only functions can be schedule into the job queue")
	ErrPeriodNotSpecified   = errors.New("unspecified job period")
	ErrParameterCannotBeNil = errors.New("nil parameters cannot be used with reflection")
)

type Job struct {
	mu *sync.RWMutex

	Func

	interval int // pause interval * unit between runs

	unit timeUnit // time units, ,e.g. 'minutes', 'hours'...

	startsImmediately bool // if the Job should run upon scheduler start

	// TODO 这个东西干嘛的??
	atTimes []time.Duration // optional time(s) at which this Job runs when interval is day

	// startAtTime time.Time // optional time at which the Job starts

	err error // error related to job

	// loc      *time.Location           // optional timezone that the atTime is in

	prevRun time.Time // datetime of the run before last run
	lastRun time.Time // datetime of last run
	nextRun time.Time // datetime of next run

	// scheduledWeekdays []time.Weekday
	// daysOfTheMonth    []int
	// tags              []string // allow the user to tag jobs with certain labels

	timer *time.Timer // handles running tasks at specific time

	// // github.com/robfig/cron/v3
	// cronSchedule cron.Schedule

	// runWithDetails bool
}

type Func struct {
	ctx    context.Context
	cancel context.CancelFunc

	isRunning *atomic.Bool
	stopped   *atomic.Bool

	fn        interface{}
	params    []interface{}
	paramsLen int

	tag      string
	funcName string

	// singletonQueueMu  *sync.Mutex
	// singletonQueue    chan struct{}
	// singletonRunnerOn *atomic.Bool

	// runStartCount *atomic.Int64
	// runFinishCount *atomic.Int64

	// singletonWg   *sync.WaitGroup
	// singletonWgMu *sync.Mutex

	fnNextRun time.Time // the next time the job is scheduled to run
}

func (f *Func) copy() (cp Func) {
	cp = Func{
		ctx:    f.ctx,
		cancel: f.cancel,

		isRunning: f.isRunning,
		stopped:   f.stopped,

		fn:        f.fn,
		params:    nil,
		paramsLen: f.paramsLen,

		tag:      f.tag,
		funcName: f.funcName,

		fnNextRun: f.fnNextRun,
	}
	cp.params = append(cp.params, f.params...)
	return
}

func newJob(interval int, startsImmediately bool) *Job {
	job := &Job{
		mu: &sync.RWMutex{},

		Func: Func{
			isRunning: atomic.NewBool(false),
			stopped:   atomic.NewBool(false),

			// runStartCount: atomic.NewInt64(0),
			// runFinishCount: atomic.NewInt64(0),
			// singletonRunnerOn: atomic.NewBool(false),
		},

		interval: interval,

		unit: seconds,

		startsImmediately: startsImmediately,

		// lastRun:  time.Unix(0, 0),
		prevRun: time.Time{},
		lastRun: time.Time{},
		nextRun: time.Time{},

		// tags: []string{},
	}

	// TODO: sched.runJobs 会重新 ctx, cancel .. 这里没必要吧??
	job.ctx, job.cancel = context.WithCancel(context.Background())

	// if singletonMode {
	//     job.SingletonMode()
	// }
	return job
}

// True if the job should be run now
func (j *Job) shouldRun() bool {
	return time.Now().Unix() >= j.nextRun.Unix()
}

// func (j *Job) neverRan() bool {
//     return j.LastRun().IsZero()
// }

func (j *Job) LastRun() time.Time {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.lastRun
}

func (j *Job) setLastRun(t time.Time) {
	j.prevRun = j.lastRun
	j.lastRun = t
}

func (j *Job) Interval() int {
	// if j.randomizeInterval {
	//     return j.getRandomInterval()
	// }
	return j.interval
}

func (j *Job) StartsImmediately() bool {
	return j.startsImmediately
}

func (j *Job) setStartsImmediately(b bool) {
	j.startsImmediately = b
}

// func (j *Job) StartAtTime() time.Time {
//     j.mu.RLock()
//     defer j.mu.RUnlock()
//     return j.startAtTime
// }

// func (j *Job) setStartAtTime(t time.Time) {
//     j.mu.Lock()
//     defer j.mu.Unlock()
//     j.startAtTime = t
// }

func (j *Job) FirstAtTime() time.Duration {
	var t time.Duration
	if len(j.atTimes) > 0 {
		t = j.atTimes[0]
	}
	return t
}

func (j *Job) LpopAtTime() time.Duration {
	var t time.Duration
	if len(j.atTimes) > 0 {
		t, j.atTimes = j.atTimes[0], j.atTimes[1:]
	}
	return t
}

func (j *Job) addAtTime(t time.Duration) {
	if len(j.atTimes) == 0 {
		j.atTimes = append(j.atTimes, t)
		return
	}

	exist := false
	index := sort.Search(len(j.atTimes), func(i int) bool {
		atTime := j.atTimes[i]
		b := atTime >= t
		if b {
			exist = atTime == t
		}
		return b
	})

	// ignore if present
	if exist {
		return
	}

	j.atTimes = append(j.atTimes, time.Duration(0))
	copy(j.atTimes[index+1:], j.atTimes[index:])
	j.atTimes[index] = t
}

// NextRun returns the time the job will run next
func (j *Job) NextRun() time.Time {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.nextRun
}

func (j *Job) setNextRun(t time.Time) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.nextRun = t
	j.Func.fnNextRun = t
}

func (j *Job) getUnit() timeUnit {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.unit
}

func (j *Job) setUnit(unit timeUnit) {
	j.mu.Lock()
	j.mu.Unlock()
	j.unit = unit
}

// TODO job 怎么会有 timer ?? 不仅仅是静态的载体?
func (j *Job) setTimer(t *time.Timer) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.timer = t
}

func (j *Job) stop() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.stopped.Store(true)

	if j.timer != nil {
		j.timer.Stop()
	}

	if j.cancel != nil {
		j.cancel()
		// TODO: 这个东西... because stop 不删除.. but 没必要吧
		// j.ctx, j.cancel = context.WithCancel(context.Background())
	}
}

func (j *Job) copy() Job {
	cp := Job{
		mu:   &sync.RWMutex{},
		Func: j.Func,
		// TODO fill fields
	}
	return cp
}

// // Run the job and immediately reschedule it
// func (j *Job) run() ([]reflect.Value, error) {
//     if j.lock {
//         if locker == nil {
//             return nil, fmt.Errorf("trying to lock %s with nil locker", j.jobFunc)
//         }
//         key := getFunctionKey(j.jobFunc)

//         locker.Lock(key)
//         defer locker.Unlock(key)
//     }
//     // TODO: 不是。。。 这个是, 本身不就是一个???
//     result, err := callJobFuncWithParams(j.funcs[j.jobFunc], j.fparams[j.jobFunc])
//     if err != nil {
//         return nil, err
//     }

//     // NOTE: reflect.ValueOf(jobFunc).Call(in); 返回 []reflect.Value, 用 interface{} / []interface{} 会有问题?? 或 有什么区别
//     return result, nil
// }

// func (j *Job) Err() error {
//     return j.err
// }

// // Do specifies the jobFunc that should be called every time the job runs
// func (j *Job) Do(jobFun interface{}, params ...interface{}) error {
//     if j.err != nil {
//         return j.err
//     }

//     typ := reflect.TypeOf(jobFun)
//     if typ.Kind() != reflect.Func {
//         return ErrNotAFunction
//     }

//     fname := getFunctionName(jobFun)
//     j.funcs[fname] = jobFun
//     j.fparams[fname] = params
//     j.jobFunc = fname

//     now := time.Now().In(j.loc)
//     // TODO: 这个是时间比较用法??
//     if !j.nextRun.After(now) {
//         j.scheduleNextRun()
//     }

//     return nil
// }

// // recovery Deprecated, 那么 panic 会影响 schedule???
// // func (j *Job) DoSafely(jobFun interface{}, params ...interface{}) error {
// //     recoveryWrapperFunc := func() {
// //         defer func() {
// //             if r := recover(); r != nil {
// //                 log.Printf("Internal panic occurred: %s", r)
// //             }
// //         }()

// //         _, _ = callJobFuncWithParams(jobFun, params)
// //     }

// //     return j.Do(recoveryWrapperFunc)
// // }

// // At schedules job at specific time of day
// //
// //	s.Every(1).Day().At("10:30:01").Do(task)
// func (j *Job) At(t string) *Job {
//     hour, min, sec, err := formatTime(t)
//     if err != nil {
//         j.err = ErrTimeFormat
//         return j
//     }

//     j.atTime = time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
//     return j
// }

// // s.Every(1).Day().At("10:30").GetAt() == "10:30"
// func (j *Job) GetAt() string {
//     return fmt.Sprintf("%d:%d", j.atTime/time.Hour, (j.atTime%time.Hour)/time.Minute)
// }

// func (j *Job) Loc(loc *time.Location) *Job {
//     j.loc = loc
//     return j
// }

// func (j *Job) Tag(t string, others ...string) {
//     j.tags = append(j.tags, t)
//     for _, tag := range others {
//         j.tags = append(j.tags, tag)
//     }
// }

// func (j *Job) Untag(t string) {
//     newTags := []string{}
//     for _, tag := range j.tags {
//         if t != tag {
//             newTags = append(newTags, tag)
//         }
//     }

//     j.tags = newTags
// }

// func (j *Job) Tags() []string {
//     return j.tags
// }

// func (j *Job) periodDuration() (time.Duration, error) {
//     interval := time.Duration(j.interval)
//     var periodDuration time.Duration

//     switch j.unit {
//     case seconds:
//         periodDuration = interval * time.Second
//     case minutes:
//         periodDuration = interval * time.Minute
//     case hours:
//         periodDuration = interval * time.Hour
//     case days:
//         periodDuration = interval * time.Hour * 24
//     case weeks:
//         periodDuration = interval * time.Hour * 24 * 7
//     default:
//         return 0, ErrPeriodNotSpecified
//     }
//     return periodDuration, nil
// }

// func (j *Job) roundToMidnight(t time.Time) time.Time {
//     return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, j.loc)
// }

// func (j *Job) scheduleNextRun() error {
//     now := time.Now()
//     if j.lastRun == time.Unix(0, 0) {
//         j.lastRun = now
//     }

//     periodDuration, err := j.periodDuration()
//     if err != nil {
//         return err
//     }

//     switch j.unit {
//     case seconds, minutes, hours:
//         j.nextRun = j.lastRun.Add(periodDuration)
//     case days:
//         j.nextRun = j.roundToMidnight(j.lastRun)
//         j.nextRun = j.nextRun.Add(j.atTime)
//     case weeks:
//         j.nextRun = j.roundToMidnight(j.lastRun)
//         dayDiff := int(j.startDay)
//         dayDiff -= int(j.nextRun.Weekday())
//         if dayDiff != 0 {
//             j.nextRun = j.nextRun.Add(time.Duration(dayDiff) * 24 * time.Hour)
//         }
//         j.nextRun = j.nextRun.Add(j.atTime)
//     }

//     // advance to next possible schedule
//     for j.nextRun.Before(now) || j.nextRun.Before(j.lastRun) {
//         j.nextRun = j.nextRun.Add(periodDuration)
//     }

//     return nil
// }

// func (j *Job) NextScheduledTime() time.Time {
//     return j.nextRun
// }

// // Lock prevents job to run from multiple instances of gocron
// func (j *Job) Lock() *Job {
//     j.lock = true
//     return j
// }
