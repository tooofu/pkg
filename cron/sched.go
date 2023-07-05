package cron

import (
	"reflect"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

// Sched struct, the only data member is the list of jobs
// - impl the sort.Interface{} for sorting jobs, by the time nextRun
type Sched struct {
	mu  sync.RWMutex
	log *logrus.Logger

	jobs []*Job

	loc *time.Location // Location to use when scheduling jobs with specified times

	running *atomic.Bool // represents if the scheduler is running at the moment or not

	execMap map[string]*executor

	// maxConcurrency int
	perConcurrency int
	mode           limitMode

	// time time.Time
	// timer func(d time.Duration, f func()) *time.Timer

	waitForInterval bool // default jobs to waiting for first interval to start
	// singletonMode   bool // default all jobs to use SingletonMode

	inSched bool
}

const (
	defaultKey = ""
)

// Len returns the number of Jobs in the Scheduler - implemented for sort
func (s *Sched) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs)
}

func (s *Sched) Swap(i, j int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[i], s.jobs[j] = s.jobs[j], s.jobs[i]
}

func (s *Sched) Less(i, j int) bool {
	// // TODO
	// // s.Jobs returns with rlock & runlock
	// return s.Jobs()[j].nextRun.Unix() >= s.Jobs()[i].nextRun.Unix()
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobs[j].NextRun().Unix() >= s.jobs[i].NextRun().Unix()
}

func (s *Sched) Jobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobs
}

func NewSched(loc *time.Location, log *logrus.Logger) (s *Sched) {
	em := make(map[string]*executor)
	em[defaultKey] = nil

	s = &Sched{
		log: log,

		jobs:    make([]*Job, 0),
		loc:     loc,
		running: atomic.NewBool(false),

		execMap: em,
		// TODO
		// time:       time.Time{}, // (0, 0) ???
		// timer: afterFunc, // ??

		// tagsUnique: false,
	}
	return
}

// func (s *Sched) SetConcurrency(n, k int, mode limitMode) {
func (s *Sched) SetConcurrency(k int, mode limitMode) {
	// s.limitMaxConcurrency = n
	s.perConcurrency = k
	s.mode = mode
}

// s.StartAysnc()
func (s *Sched) StartAsync() {
	if !s.IsRunning() {
		s.start()
	}
}

func (s *Sched) start() {
	s.setRunning(true)

	// TODO  通过 chan 与 job / sched 关联..
	// TODO sync.WaitGroup, stop sched wait all executor done
	for k := range s.execMap {
		e := newExecutor(s.log)
		e.limitPerConcurrency = s.perConcurrency
		e.start()

		s.execMap[k] = e
	}
	// for i := 1; i < s.limitMaxConcurrency; i++ {
	//     s.executor.start()
	// }

	s.runJobs(s.Jobs())
}

// IsRunning returns true if the scheduler is running
func (s *Sched) IsRunning() bool {
	return s.running.Load()
}

func (s *Sched) setRunning(b bool) {
	s.running.Store(b)
}

func (s *Sched) runJobs(jobs []*Job) {
	for _, job := range jobs {
		// // TODO: pause 时, 先通过 cancel .. 再 ??
		// ctx, cancel := context.WithCancel(context.Background())
		// job.mu.Lock()
		// job.ctx = ctx
		// job.cancel = cancel
		// job.mu.Unlock()

		s.runContinue(job)
	}
}

func (s *Sched) runContinue(job *Job) {
	should, dur := s.schedNextRun(job)
	// s.log.Debugln(time.Now(), "before", should, dur, job.StartsImmediately(), len(s.jobs))
	if !should {
		return
	}

	// NOTE: startsImmediately
	// 1. 作为 switch, 初始化 / 某种 sched 可能出现 false
	// 2. 通过初始化时的 time.AfterFunc + startsImmediately=true, 直接 s.run(job)
	if !job.StartsImmediately() {
		job.setStartsImmediately(true)
	} else {
		s.run(job)
	}

	// s.log.Debugln(time.Now(), "after", should, dur, job.StartsImmediately(), s.IsRunning(), len(s.jobs))

	// nr := next.dateTime.Sub(s.now())
	// if nr < 0 {
	//     job.setLastRun(s.now())
	//     shouldRun, next := s.schedNextRun(job)
	//     if !shouldRun {
	//         return
	//     }
	//     nr = next.dateTime.Sub(s.now())
	// }

	// TODO remove
	// if job.timer != nil {
	//     return
	// }
	// fmt.Printf("job:%v,dur:%#v\n", job, dur)

	// NOTE
	// job 的 perputual
	// 1.1 不是 sched 的一个大 timer 通过 sort.Sort 排序 sched.jobs slice 来, 并 sleep top - now() 来实现 perputual
	// 1.2 wakeup 的控制权原本是 for + time.Sleep(top - now()) & run ready jobs(sorts and each job.nextRun.before(now))
	// 2.1 wakeup 的控制权, 全全交给 runtime.timer ... 不参与.. 只依赖 time.AfterFunc(每次设置下一次) ..
	job.setTimer(time.AfterFunc(dur, func() {
		// // TODO: 这个东西是用来停止的??
		// if !next.dateTime.IsZero() {
		//     for {
		//         n := s.now().UnixNano() - next.dateTime.UnixNano()
		//         if n >= 0 {
		//             break
		//         }

		//         // 这个没有 body 的.. 防止 老 for ??
		//         select {
		//         case <-s.executor.ctx.Done():
		//         case <-time.After(time.Duration(n)):
		//         }
		//     }
		// }

		s.runContinue(job)
	}))
}

func (s *Sched) run(job *Job) {
	if !s.IsRunning() {
		return
	}

	job.mu.Lock()
	if job.fn == nil {
		job.mu.Unlock()
		s.RemoveJob(job)
		return
	}
	defer job.mu.Unlock()

	// if job.runWithDetails {
	//     switch len(job.params) {
	//     case job.paramsLen:
	//         job.params = append(job.params, job.copy())
	//     case job.paramsLen + 1:
	//         job.params[job.paramsLen] = job.copy()
	//     default:
	//         // something is really wrong and we should never get here
	//         job.err = wrapOrError(job.err, ErrInvaildFunctionParameters)
	//         return
	//     }
	// }

	// 这步是 already job, 只剩 executor 执行
	// TODO: 调度与并发??
	// 1. 调度, 通过 job.tag 分发到不同 ch ?  / 不同 executor ?
	// 2. s.executor 的 size ?

	switch s.mode {
	case IgnoreMode:
		select {
		case s.execMap[job.tag].jobFnCh <- job.Func.copy():
		default:
			// TODO ignore count
		}
	case WaitMode:
		s.execMap[job.tag].jobFnCh <- job.Func.copy()
	}
}

// func (s *Sched) schedNextRun(job *Job) (bool, nextTime) {
func (s *Sched) schedNextRun(job *Job) (should bool, dur time.Duration) {
	// TODO: sched.mu.Lock 更新 jobs 这个有必要??
	// if !s.jobPresent(job) {
	//     return false, nextRun{}
	// }

	if !job.shouldRun() {
		s.RemoveJob(job)
		return
	}

	nextRun := job.NextRun()

	dur = s.durationToNext(nextRun, job)

	// // next time
	// nt := s.durationToNext(nextRun, job)

	// now := s.now()
	// // TODO: 为什么这里设置 lastRun ???
	// jobNextRun := job.NextRun()
	// if jobNextRun.After(now) {
	//     job.setLastRun(now)
	// } else {
	//     job.setLastRun(jobNextRun)
	// }

	// if nt.dateTime.IsZero() {
	//     nt.dateTime = lastRun.Add(next.duration)
	//     job.setNextRun(next.dateTime)
	// } else {
	//     job.setNextRun(next.dateTime)
	// }

	return true, dur
}

func (s *Sched) now() time.Time {
	return time.Now().In(s.loc)
}

// func (s *Sched) jobPresent(j *Job) bool {
//     // treemap??
//     // concat(time.Time, seen)
//     // map with impl sort
//     s.mu.RLock()
//     defer s.mu.RUnlock()
//     for _, job := range s.Jobs() {
//         if job == j {
//             return true
//         }
//     }
//     return false
// }

// durationToNext calculate how much time to the next run, depending on unit
func (s *Sched) durationToNext(nextRun time.Time, job *Job) (dur time.Duration) {
	// func (s *Sched) durationToNext(nextRun time.Time, job *Job) (nt nextTime) {
	unit := job.getUnit()
	interval := job.Interval()
	// NOTE 只支持, day, hour, minute, second.. 相对高频, week, month, year, 暂不考虑
	switch unit {
	case seconds:
		dur = time.Duration(interval) * time.Second
	case minutes:
		dur = time.Duration(interval) * time.Minute
	case hours:
		dur = time.Duration(interval) * time.Hour
	case days:
		if interval == 1 {
			// dur = job.FirstAtTime()
			dur = job.LpopAtTime()
		}
		//     nt.dateTime = s.calculateDays(job, lastRun)
		// dur = s.calculateDays(job, lastRun)
	}

	return
}

// func (s *Sched) calculateDays(job *Job, lastRun time.Time) nextRun {
//     if job.getInterval() == 1 {
//         lastRunDayPlusJobAtTime := s.roundToMidnightAndAddDSTAware(lastRun, job.getAtTime(lastRun))

//         if shouldRunToday(lastRun, lastRunDayPlusJobAtTime) {
//             return nextRun{
//                 duration: until(lastRun, lastRunDayPlusJobAtTime),
//                 dateTime: lastRunDayPlusJobAtTime,
//             }
//         }
//     }

//     nextRunAtTime := s.roundToMidnightAndAddDSTAware(lastRun, job.getFirstAtTime()).AddDate(0, 0, job.getInterval()).In(s.loc)
//     return nextRun{
//         duration: until(lastRun, nextRunAtTime),
//         dateTime: nextRunAtTime,
//     }
// }

func (s *Sched) Clear() {
}

func (s *Sched) Stop() {
	if s.IsRunning() {
		s.stop()
	}
}

func (s *Sched) stop() {
	for _, job := range s.jobs {
		job.stop()
	}
	for _, executor := range s.execMap {
		executor.stop()
	}
	s.setRunning(false)
}

// 1
// s.Every(5).Seconds().Do(func(){ ... })
// s.Every(5).Days().Do(func(){ ... })
// 2
// s.Every(1).Month(1, 2, 3).Do(func(){ ... })
// s.Every(1).Day().At("10:30").Do(func(){ ... })
// s.Every(1).Day().At("10:30;08:00").Do(func(){ ... })
// 3
// s.Every(1).Day().At("10:30-12:00;13:00-18:00").Do(func(){ ... })
// 4
// // strings parse to duration
// s.Every("5m").Do(func(){ ... })
func (s *Sched) Every(interval interface{}) *Sched {
	job := s.getCurrJob()

	switch interval := interval.(type) {
	case int:
		job.interval = interval
		if interval <= 0 {
			job.err = wrapOrError(job.err, ErrInvalidInterval)
		}
	default:
		job.err = wrapOrError(job.err, ErrInvalidIntervalType)
	}
	return s
}

func (s *Sched) getCurrJob() *Job {
	if !s.inSched {
		s.mu.Lock()
		s.jobs = append(s.jobs, s.newJob(0))
		s.mu.Unlock()
		s.inSched = true
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// job := NewJob(interval).Loc(s.loc)
	return s.jobs[len(s.jobs)-1]
}

func (s *Sched) newJob(interval int) *Job {
	return newJob(interval, !s.waitForInterval)
}

func (s *Sched) setUnit(unit timeUnit) {
	job := s.getCurrJob()
	// currUnit := job.getUnit()
	// if currUnit == duration || currUnit == crontab {
	//     job.err = wrapOrError(job.err, ErrInvalidIntervalUnitsSelection)
	// }
	job.setUnit(unit)
}

func (s *Sched) Day() *Sched {
	return s.Days()
}

func (s *Sched) Days() *Sched {
	s.setUnit(days)
	return s
}

func (s *Sched) Hour() *Sched {
	return s.Hours()
}

func (s *Sched) Hours() *Sched {
	s.setUnit(hours)
	return s
}

func (s *Sched) Minute() *Sched {
	return s.Minutes()
}

func (s *Sched) Minutes() *Sched {
	s.setUnit(minutes)
	return s
}

func (s *Sched) Second() *Sched {
	return s.Seconds()
}

func (s *Sched) Seconds() *Sched {
	s.setUnit(seconds)
	return s
}

func (s *Sched) Immediately(immediately bool) *Sched {
	job := s.getCurrJob()
	job.startsImmediately = immediately
	return s
}

func (s *Sched) At(i interface{}) *Sched {
	job := s.getCurrJob()

	switch t := i.(type) {
	case string:
		// for _, tt := range strings.Split(t, ";") {
		//     hour, min, sec, err := parseTime(tt)
		hour, min, sec, err := parseTime(t)
		if err != nil {
			job.err = wrapOrError(job.err, err)
			return s
		}

		now := s.now()
		hour = now.Hour() - hour
		min = now.Minute() - min
		sec = now.Second() - sec

		dur := time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
		if dur < 0 {
			dur = dur + time.Duration(24)*time.Hour
			// NOTE: At(...).Immediately(true).Do(...)
			// if job.StartsImmediately() {
			//     job.addAtTime(time.Duration(0))
			// }
		}
		// dur = time.Duration(24) * time.Hour
		job.addAtTime(dur)
		// }
	default:
		job.err = wrapOrError(job.err, ErrUnsupportedTimeFormat)
	}

	job.startsImmediately = false
	return s
}

func (s *Sched) Tag(tag string) *Sched {
	job := s.getCurrJob()
	job.tag = tag

	// TODO with mutex?
	if _, ok := s.execMap[tag]; !ok {
		s.execMap[tag] = nil
	}
	return s
}

func (s *Sched) Do(fn interface{}, params ...interface{}) (*Job, error) {
	job := s.getCurrJob()
	// NOTE: sched 的最后一步, 不再有 sched 的链式调用, 设置 false, 下一个 sched 会 new job
	s.inSched = false

	// TODO: check errors
	// jobUnit := job.getUnit()
	// jobLastRun := job.LastRun()

	// // if job.getAtTime(jobLastRun) != 0 && (jobUnit <= hours || jobUnit >= duration) {
	// if job.getAtTime(jobLastRun) != 0 && jobUnit <= hours {
	//     job.err = wrapOrError(job.err, ErrAtTimeNotSupported)
	// }

	// if len(job.schedWeekdays) != 0 && jobUnit != weeks {
	// }

	// if job.unit != crontab && job.getInterval() == 0 {
	//     if job.unit != duration {
	//         job.err = wrapOrError(job.err, ErrInvalidInterval)
	//     }
	// }

	// if job.err != nil {
	//     s.RemoveByReference(job)
	//     return nil, job.err
	// }

	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		// TODO: 这个东西... 意思 fn 是个 瞎比 interface, 非 func 剔除
		// s.RemoveByRef(job)
		return nil, ErrNotAFunction
	}

	funcName := getFunctionName(fn)
	// if job.tag == "" {
	//     job.tag = funcName
	// }
	// TODO: job 结束以后, 会复用这个坑? 还是去掉??
	// if job.funcName != funcName {
	//     job.fn = fn
	//     job.params = params
	//     job.funcName = fname
	// }
	job.fn = fn
	job.params = params
	job.funcName = funcName

	// f := reflect.ValueOf(jobFun)
	// expectedParamLength := f.Type().NumIn()
	// if job.runWithDetails {
	//     expectedParamLength--
	// }

	// if len(params) != expectedParamLength {
	//     s.RemoveByReference(job)
	//     job.err = wrapOrError(job.err, ErrWrongParams)
	//     return nil, job.err
	// }

	// if job.runWithDetails && f.Type().In

	// NOTE: 非启动时冷加载job, 需要check run
	if s.IsRunning() {
		s.runContinue(job)
	}

	return job, nil
}

func (s *Sched) WatchDog() {
	s.log.Debugf("[watch dog] len(jobs)=%d", len(s.jobs))
	for k, e := range s.execMap {
		s.log.Debugf("[watch dog] execMap k=%v len(e.jobFnCh)=%d", k, len(e.jobFnCh))
	}
}

func (s *Sched) RemoveJob(job interface{}) {
	funcName := getFunctionName(job)
	// j := s.fundJobByTaskName(fName)
	// s.removeJobsUniqueTags(j)
	s.removeByCond(func(job *Job) bool {
		return job.funcName == funcName
	})
}

// TODO 这个东西 只能是 sched 预留 tag(string) chan, svc 往里面塞...
func (s *Sched) RemoveByTag(tag string) {
	s.removeByCond(func(job *Job) bool {
		return job.tag == tag
	})
}

func (s *Sched) removeByCond(should func(*Job) bool) {
	retain := make([]*Job, 0)
	for _, job := range s.Jobs() {
		if !should(job) {
			retain = append(retain, job)
		} else {
			job.stop()
		}
	}
	s.setJobs(retain)
}

func (s *Sched) setJobs(jobs []*Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = jobs
}

// // TODO: 为啥这么写???
// func (s *Sched) getRunnableJobs() (runningJobs [MAXJOBNUM]*Job, n int) {
//     runnableJobs := [MAXJOBNUM]*Job{}
//     n = 0
//     sort.Sort(s)
//     for i := 0; i < s.size; i++ {
//         // TODO: shouldRun -> ready
//         if s.jobs[i].shouldRun() {
//             runnableJobs[n] = s.jobs[i]
//             n++
//         } else {
//             break
//         }
//     }
//     return runnableJobs, n
// }

// func (s *Sched) NextRun() (*Job, time.Time) {
//     if s.size <= 0 {
//         return nil, time.Now()
//     }
//     sort.Sort(s)
//     return s.jobs[0], s.jobs[0].nextRun
// }

// func (s *Sched) RunPending() {
//     runnableJobs, n := s.getRunnableJobs()

//     if n != 0 {
//         for i := 0; i < n; i++ {
//             go runnableJobs[i].run()
//             runnableJobs[i].lastRun = time.Now()
//             runnableJobs[i].scheduleNextRun()
//         }
//     }
// }

// func (s *Sched) RunAll() {
//     s.RunAllwithDelay(0)
// }

// func (s *Sched) RunAllwithDelay(d int) {
//     for i := 0; i < s.size; i++ {
//         go s.jobs[i].run()
//         if 0 != d {
//             time.Sleep(time.Duration(d))
//         }
//     }
// }

// func (s *Sched) Remove(j interface{}) {
//     s.delByCond(func(sb *Job) bool {
//         return sb.jobFunc == getFunctionName(j)
//     })
// }

// func (s *Sched) RemoveByRef(j *Job) {
//     s.delByCond(func(sb *Job) bool {
//         return sb == j
//     })
// }

// func (s *Sched) RemoveByTag(t string) {
//     s.delByCond(func(sb *Job) bool {
//         for _, a := range sb.tags {
//             if a == t {
//                 return true
//             }
//         }
//         return false
//     })
// }

// func (s *Sched) delByCond(handle func(*Job) bool) {
//     i := 0

//     // keep deleting until no more jobs match the criteria
//     for {
//         found := false

//         for ; i < s.size; i++ {
//             if handle(s.jobs[i]) {
//                 found = true
//                 break
//             }
//         }

//         if !found {
//             return
//         }

//         // NOTE: 这个 move 有点夸张... slice a[0:10] = a[1:11] 会怎么样, 假删会怎么样?
//         for j := i + 1; j < s.size; j++ {
//             s.jobs[i] = s.jobs[j]
//             i++
//         }
//         s.size--
//         s.jobs[s.size] = nil
//     }
// }

// func (s *Sched) Scheduled(j interface{}) bool {
//     for _, job := range s.jobs {
//         if job.jobFunc == getFunctionName(j) {
//             return true
//         }
//     }
//     return false
// }

// // Start all the pending jobs
// // Add seconds ticker
// func (s *Sched) Start() chan bool {
//     stopped := make(chan bool, 1)
//     ticker := time.NewTicker(1 * time.Second)

//     go func() {
//         for {
//             select {
//             case <-ticker.C:
//                 s.RunPending()
//             case <-stopped:
//                 ticker.Stop()
//                 // TODO remove
//                 fmt.Println("sched ticker stopped")
//                 return
//             }
//         }
//     }()

//     return stopped
// }

// var (
//     // TODO time.UTC ???
//     defaultScheduler = NewSched(time.UTC)
// )

// func Every(interval uint64) *Job {
//     return defaultScheduler.Every(interval)
// }

// func RunPending() {
//     defaultScheduler.RunPending()
// }

// func RunAll() {
//     defaultScheduler.RunAll()
// }

// func RunAllwithDelay(d int) {
//     defaultScheduler.RunAllwithDelay(d)
// }

// func Start() chan bool {
//     return defaultScheduler.Start()
// }
