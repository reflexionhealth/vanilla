package clock

import (
	"sync"
	"time"
)

func AtTime(t time.Time, block func())       { Default.AtTime(t, block) }
func In(loc *time.Location) time.Time        { return Default.In(loc) }
func UTC() time.Time                         { return Default.UTC() }
func After(d time.Duration) <-chan time.Time { return Default.After(d) }
func Tick(d time.Duration) <-chan time.Time  { return Default.Tick(d) }
func Sleep(d time.Duration)                  { Default.Sleep(d) }

type Source struct {
	Now    time.Time
	Frozen bool
	sync.Mutex
}

var Default Source

func (s Source) AtTime(t time.Time, block func()) {
	s.Lock()
	defer func() {
		s.Frozen = false
		s.Unlock()
	}()

	s.Now = t
	s.Frozen = true
	block()
}

func (s Source) In(loc *time.Location) time.Time {
	if s.Frozen {
		return s.Now
	} else {
		return time.Now().In(loc)
	}
}

func (s Source) UTC() time.Time {
	return s.In(time.UTC)
}

func (s Source) After(d time.Duration) <-chan time.Time {
	if s.Frozen {
		panic("vanilla/clock: clock.After() has not been implemented")
	} else {
		return time.After(d)
	}
}

func (s Source) Tick(d time.Duration) <-chan time.Time {
	if s.Frozen {
		panic("vanilla/clock: clock.Tick() has not been implemented")
	} else {
		return time.Tick(d)
	}
}

func (s Source) Sleep(d time.Duration) {
	if s.Frozen && d > 0 {
		panic("vanilla/clock: clock.Sleep() has not been implemented")
	} else {
		time.Sleep(d)
	}
}
