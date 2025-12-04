package task

import (
	"fmt"
	"time"
)

type Item interface {
	// Enabled 返回是否需要执行任务
	Enabled() bool
	// BeforeRun 在任务执行前调用
	BeforeRun(since time.Time) error
	// AfterRun 在任务执行后调用
	AfterRun(since time.Time, elapsed time.Duration) error
}

// Store 用于查询/存储任务的状态
type Store interface {
	// GetItem 获取任务的状态
	GetItem(name string) (item Item, found bool, err error)
	// NewItem 当任务不存在时调用该方法创建任务
	NewItem(name string) (item Item, err error)
	// SetItem 更新任务的状态
	SetItem(name string, item Item) error
}

func WithStore(store Store) Option {
	return func(task *Task) {
		wrappedFn := task.Fn
		task.Fn = func() {
			item, found, err := store.GetItem(task.Name)
			if err != nil {
				panic(fmt.Sprintf("cannot get item %q: %v", task.Name, err))
			}
			if !found {
				task.Logger.Infof("item %q not found", task.Name)
				item, err = store.NewItem(task.Name)
				if err != nil {
					panic(fmt.Sprintf("cannot create item %q: %v", task.Name, err))
				}
			}
			if !item.Enabled() {
				task.Logger.Debugf("item %q disabled", task.Name)
				return
			}

			start := time.Now()
			if err := item.BeforeRun(start); err != nil {
				panic(fmt.Errorf("cannot run item.BeforeRun %q: %v", task.Name, err))
			}
			wrappedFn()
			if err := item.AfterRun(start, time.Since(start)); err != nil {
				panic(fmt.Errorf("cannot run item.AfterRun %q: %v", task.Name, err))
			}

			if err := store.SetItem(task.Name, item); err != nil {
				panic(fmt.Errorf("cannot set item %q: %v", task.Name, err))
			}
		}
	}
}
