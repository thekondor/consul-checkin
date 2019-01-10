package watchdog

import (
	. "github.com/thekondor/consul-checkin/time"
)

type ConnectionHealthCheck interface {
	Ping() error
}

type ConnectionWatchdog struct {
	healthCheck ConnectionHealthCheck
	options     *WatchOptions
	pingTicker  Ticker
	ch          struct {
		onStop    chan struct{}
		watchDone chan struct{}
	}
}

func (self *ConnectionWatchdog) Stop() {
	self.pingTicker.Stop()

	close(self.ch.onStop)
	<-self.ch.watchDone
}

func (self *ConnectionWatchdog) ping() error {
	return self.healthCheck.Ping()
}

func (self *ConnectionWatchdog) watchConnection(initialErr error) {
	var (
		lastErr error = initialErr
		attempt uint  = 0
	)

event_loop:
	for {
		select {
		case <-self.pingTicker.Ch():
			err := self.ping()
			if nil == err {
				if nil != lastErr {
					lastErr = nil
					self.options.On.ConnectionRecovered()
				}

				continue event_loop
			}

			if nil != lastErr {
				lastErr = err
				if self.tobeContinuedOnFailedConnection(attempt, err) {
					attempt++
					continue
				}

				break event_loop
			} else {
				if self.tobeContinuedOnLostConnection(err) {
					attempt++
					continue
				}
				break event_loop
			}

		case <-self.ch.onStop:
			break event_loop
		}
	}

	close(self.ch.watchDone)
}

func (self *ConnectionWatchdog) tobeContinuedOnFailedConnection(attempt uint, err error) bool {
	return self.options.On.ConnectionFailed(attempt, err).bool()
}

func (self *ConnectionWatchdog) tobeContinuedOnLostConnection(err error) bool {
	return self.options.On.ConnectionLost(err).bool()
}

func WatchConnection(healthCheck ConnectionHealthCheck, options *WatchOptions) (ConnectionWatchdog, error) {
	options.adjust()

	instance := ConnectionWatchdog{
		healthCheck: healthCheck,
		options:     options,
		pingTicker:  options.NewTicker(options.PingInterval),
	}
	instance.ch.onStop = make(chan struct{})
	instance.ch.watchDone = make(chan struct{})

	initialPingErr := instance.ping()
	go instance.watchConnection(initialPingErr)

	return instance, initialPingErr
}
