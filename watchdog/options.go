package watchdog

import (
	. "github.com/thekondor/consul-checkin/time"
	"time"
)

type WatchOptions struct {
	PingInterval time.Duration
	On           ConnectionEvents
	// TODO: replace PingInterval + NewTicker with a single object
	NewTicker NewTickerFunc
}

type ConnectionEvents struct {
	ConnectionFailed    OnConnectionFailedFunc
	ConnectionLost      OnConnectionLostFunc
	ConnectionRecovered OnConnectionRecoveredFunc
}

func (options *WatchOptions) adjust() {
	if nil == options.On.ConnectionFailed {
		options.On.ConnectionFailed = func(uint, error) RetryDecision {
			return GiveUp
		}
	}
	if nil == options.On.ConnectionLost {
		options.On.ConnectionLost = func(error) RetryDecision {
			return GiveUp
		}
	}
	if nil == options.On.ConnectionRecovered {
		options.On.ConnectionRecovered = func() {}
	}
	if 0 == options.PingInterval {
		options.PingInterval = 3 * time.Second
	}

	if nil == options.NewTicker {
		options.NewTicker = NewClockTicker
	}
}
