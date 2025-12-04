package task

import (
	"context"
	"fmt"
	"github.com/gorhill/cronexpr"
	"github.com/gridexswap/utils/log"
	"io"
	"runtime/debug"
	"time"
)

const (
	CronEverySecond    = "* * * * * * *"
	CronEveryMinute    = "* * * * * *"
	CronEvery10Minutes = "*/10 * * * * *"
	CronEveryHour      = "0 0/1 * * * *"
	CronEveryDay       = "0 0 * * * *"
)

var (
	_ io.Closer = (*Task)(nil)
	_ Runner    = (*Task)(nil)
)

type Runner interface {
	Run()
}

type Option func(*Task)

type Task struct {
	ticker       *time.Ticker
	interval     time.Duration
	initialDelay time.Duration

	cronExpr *cronexpr.Expression

	ctx    context.Context
	cancel context.CancelFunc // cancel func for ctx

	Logger log.Logger
	Name   string // task name
	Fn     func()
}

// Run runs the task
//
// If Run function finishes, it never closes the task
func (t *Task) Run() {
	if t.cronExpr == nil && t.initialDelay >= 0 {
		initialDelayTimer := time.NewTimer(t.initialDelay)

		// 如果设置了context，需要检查是否已经取消
		select {
		case <-t.ctx.Done():
			initialDelayTimer.Stop()
			return
		case <-initialDelayTimer.C:
			initialDelayTimer.Stop()
			t.Fn()
		}
	}

	// 如果是cron task，则需要计算首次运行的时间间隔
	interval := t.interval
	if t.cronExpr != nil {
		fromTime := time.Now()
		nextTime := t.cronExpr.Next(fromTime)
		interval = nextTime.Sub(fromTime)
	}

	t.ticker = time.NewTicker(interval)

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-t.ticker.C:
			if t.cronExpr != nil {
				fromTime := time.Now()
				nextTime := t.cronExpr.Next(fromTime)
				t.ticker.Reset(nextTime.Sub(fromTime))
			}

			t.Fn()
		}
	}
}

// RunAndClose runs the task and closes it
//
// If Run function finishes, it will close the task
func (t *Task) RunAndClose() {
	defer t.Close()

	t.Run()
}

func (t *Task) Close() error {
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.cancel()
	return nil
}

// NewTask creates a task with fixed interval
func NewTask(interval time.Duration, fn func(), opts ...Option) *Task {
	return newTask(interval, nil, fn, opts...)
}

// NewCronTask creates a cron task
//
// cron expression format: https://en.wikipedia.org/wiki/Cron#CRON_expression
func NewCronTask(cron string, fn func(), opts ...Option) (*Task, error) {
	expr, err := cronexpr.Parse(cron)
	if err != nil {
		return nil, fmt.Errorf("parse cron expression %s error: %w", cron, err)
	}
	return newTask(time.Duration(0), expr, fn, opts...), nil
}

func MustNewCronTask(cron string, fn func(), opts ...Option) *Task {
	task, err := NewCronTask(cron, fn, opts...)
	if err != nil {
		panic(err)
	}
	return task
}

func newTask(interval time.Duration, cronExpr *cronexpr.Expression, fn func(), opts ...Option) *Task {
	t := &Task{
		interval:     interval,
		initialDelay: -1, // 不需要首次运行
		cronExpr:     cronExpr,
		Logger:       log.DefaultLogger,
		Fn:           fn,
		Name:         "unknown",
	}
	for _, opt := range opts {
		opt(t)
	}

	// 如果没有设置context，则使用默认的context
	if t.cancel == nil {
		t.ctx, t.cancel = context.WithCancel(context.Background())
	}

	if cronExpr == nil {
		// 配置的initial delay比interval大，不需要首次运行
		if t.initialDelay >= 0 && t.initialDelay >= interval {
			t.initialDelay = -1
		}
	}

	return t
}

// WithName sets the task Name
func WithName(name string) Option {
	return func(task *Task) {
		task.Name = name
	}
}

// WithContext sets the context
//
// If not set, the default context.Background() will be used. It can be used to cancel the task.
// For example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go taskutil.NewTask(interval, fn, taskutil.WithContext(ctx)).Run()
func WithContext(parent context.Context) Option {
	return func(task *Task) {
		task.ctx, task.cancel = context.WithCancel(parent)
	}
}

// WithLogger sets the Logger
func WithLogger(logger log.Logger) Option {
	return func(task *Task) {
		task.Logger = logger
	}
}

// WithPanicRecover recovers from panic
func WithPanicRecover() Option {
	return func(task *Task) {
		wrappedFn := task.Fn
		task.Fn = func() {
			defer func() {
				if r := recover(); r != nil {
					task.Logger.Errorf("task %s panic recovered: %v, stack: \n%s", task.Name, r, debug.Stack())
				}
			}()

			wrappedFn()
		}
	}
}

// WithElapsed measures the time elapsed for the task execution
func WithElapsed() Option {
	return func(task *Task) {
		wrappedFn := task.Fn

		task.Fn = func() {
			start := time.Now()
			defer func() {
				task.Logger.Infof("execute task \"%s\" with time elapsed: %v", task.Name, time.Since(start))
			}()

			wrappedFn()
		}
	}
}

// WithInitialDelay 设置首次运行的延迟时间
//
// 当首次运行的延迟时间大于等于interval时，或delay小于0时，不需要首次运行
//
// 如果为cron task，不支持设置首次运行的延迟时间
func WithInitialDelay(delay time.Duration) Option {
	return func(task *Task) {
		if task.cronExpr != nil {
			panic("cannot set initial delay for cron task")
		}

		task.initialDelay = delay
	}
}
