package time

import "time"

type NewTickerFunc func(d time.Duration) Ticker

type Ticker interface {
	Ch() <-chan time.Time
	Stop()
}

type ClockTicker struct {
	*time.Ticker
}

func NewClockTicker(duration time.Duration) Ticker {
	ticker := ClockTicker{time.NewTicker(duration)}
	return &ticker
}

func (self *ClockTicker) Ch() <-chan time.Time {
	return self.C
}

func (self *ClockTicker) Stop() {
	self.Ticker.Stop()
}
