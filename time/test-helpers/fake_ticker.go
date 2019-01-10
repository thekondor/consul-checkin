package test_helpers

import (
	"time"
)

type FakeTicker struct {
	ch chan time.Time
}

func NewFakeTicker(_ time.Duration) *FakeTicker {
	return &FakeTicker{ch: make(chan time.Time, 1)}
}

func (ft *FakeTicker) Ch() <-chan time.Time {
	return ft.ch
}

func (ft *FakeTicker) Stop() {}

func (ft *FakeTicker) Tick() {
	ft.ch <- time.Now()
}

func (ft *FakeTicker) TickTimes(times int) {
	for attempt := 0; attempt < times; attempt++ {
		ft.Tick()
	}
}
