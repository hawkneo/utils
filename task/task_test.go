package task

import (
	"context"
	"github.com/gorhill/cronexpr"
	"github.com/gridexswap/utils/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	t.Run("without panic", func(t *testing.T) {
		i := 3
		ta := NewTask(200*time.Millisecond, func() {
			t.Log("hello")
			i--
			if i <= 0 {
				t.SkipNow()
			}
		})
		defer ta.Close()

		ta.Run()
	})

	t.Run("with panic", func(t *testing.T) {
		i := 3
		ta := NewTask(200*time.Millisecond, func() {
			i--
			if i < 0 {
				t.SkipNow()
			}
			panic("panic")
		}, WithPanicRecover())
		defer ta.Close()

		ta.Run()
	})

	t.Run("with options", func(t *testing.T) {
		i := 3
		ta := NewTask(200*time.Millisecond, func() {
			t.Log("hello")
			i--
			if i <= 0 {
				t.SkipNow()
			}
		},
			WithName("my task name"),
			WithLogger(log.AnsiColorLogger{ColorOutput: true}),
			WithPanicRecover(),
			WithElapsed(),
		)
		defer ta.Close()

		ta.Run()
	})
}

func TestNewTask_WithContext(t *testing.T) {
	t.Run("run with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		ta := NewTask(200*time.Millisecond, func() {
			t.FailNow()
		},
			WithContext(ctx),
			WithInitialDelay(time.Millisecond),
		)
		defer ta.Close()

		ta.Run()
	})

	t.Run("run with cancellable context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		i := 3
		ta := NewTask(200*time.Millisecond, func() {
			t.Log("hello")
			i--
			if i <= 0 {
				cancel()
			}
		},
			WithContext(ctx),
		)
		defer ta.Close()

		ta.Run()
	})
}

func TestNewCronTask(t *testing.T) {
	fromTime := time.Now()
	expr := cronexpr.MustParse(CronEveryMinute)
	times := expr.NextN(fromTime, 5)
	for _, ti := range times {
		t.Log("time", ti)
	}
	i := 0
	ta, err := NewCronTask(CronEverySecond, func() {
		i++
		t.Log("time", time.Now())
		if i > 3 {
			t.SkipNow()
		}
	})
	require.NoError(t, err)
	defer ta.Close()

	ta.Run()
}

func TestTask_Close(t *testing.T) {
	t.Run("close before initial delay reached", func(t *testing.T) {
		ta := NewTask(200*time.Millisecond, func() {
			t.FailNow()
		},
			WithInitialDelay(100*time.Millisecond),
		)
		require.NoError(t, ta.Close())

		ta.Run()
	})

	t.Run("close after initial delay reached", func(t *testing.T) {
		i := 0
		var ta *Task
		ta = NewTask(200*time.Millisecond, func() {
			i++
			require.NoError(t, ta.Close())
		},
			WithInitialDelay(100*time.Millisecond),
		)
		ta.Run()
	})
}

func TestWithStore(t *testing.T) {
	t.Run("run with enabled", func(t *testing.T) {
		i := 3
		ta := NewTask(200*time.Millisecond, func() {
			t.Log("hello")
			i--
			if i <= 0 {
				t.SkipNow()
			}
		},
			WithName("my task name"),
			WithLogger(log.AnsiColorLogger{ColorOutput: true}),
			WithPanicRecover(),
			WithElapsed(),
			WithStore(&inMemoryStore{items: make(map[string]*item), enabled: true}),
		)
		ta.Run()
	})
}

var _ Store = (*inMemoryStore)(nil)
var _ Item = (*item)(nil)

type item struct {
	enabled bool
}

func (i item) Enabled() bool {
	return i.enabled
}

func (i item) BeforeRun(time.Time) error {
	return nil
}

func (i item) AfterRun(time.Time, time.Duration) error {
	return nil
}

type inMemoryStore struct {
	items   map[string]*item
	enabled bool
}

func (store *inMemoryStore) GetItem(name string) (item Item, found bool, err error) {
	item, found = store.items[name]
	return
}

func (store *inMemoryStore) NewItem(name string) (i Item, err error) {
	is := &item{enabled: store.enabled}
	store.items[name] = is
	i = is
	return
}

func (store *inMemoryStore) SetItem(name string, i Item) error {
	store.items[name] = i.(*item)
	return nil
}
