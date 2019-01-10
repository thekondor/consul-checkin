package watchdog

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	. "github.com/thekondor/consul-checkin/time"
	. "github.com/thekondor/consul-checkin/time/test-helpers"
	. "github.com/thekondor/consul-checkin/watchdog/test-helpers"
	"log"
	"testing"
	"time"
)

const InvalidRetryDecision RetryDecision = RetryDecision(42)

type ConnectionWatchdogTest struct {
	suite.Suite

	healthCheckMock *ConnectionHealthCheckMock
	fakeTicker      *FakeTicker
	assert          *assert.Assertions
	require         *require.Assertions
}

func TestConnectionWatchdog(t *testing.T) {
	suite.Run(t, new(ConnectionWatchdogTest))
}

func (test *ConnectionWatchdogTest) returnFakeTicker(_ time.Duration) Ticker {
	return test.fakeTicker
}

func (test *ConnectionWatchdogTest) Logf(format string, args ...interface{}) {
	log.Printf("[test] "+format, args...)
}

func (test *ConnectionWatchdogTest) Log(args ...interface{}) {
	args = append([]interface{}{"[test]"}, args...)
	log.Println(args...)
}

func (test *ConnectionWatchdogTest) SetupTest() {
	test.fakeTicker = NewFakeTicker(0 * time.Second)
	test.healthCheckMock = new(ConnectionHealthCheckMock)

	test.assert = assert.New(test.T())
	test.require = require.New(test.T())
}

func (test *ConnectionWatchdogTest) Test_ContinuesToPing_WhenConnectionIsRecovered_OnBrokenInitially() {
	test.healthCheckMock.On("Ping").
		Return(errors.New("broken connection")).
		Once()

	sut, err := WatchConnection(test.healthCheckMock, &WatchOptions{
		NewTicker: test.returnFakeTicker,
		On: ConnectionEvents{
			ConnectionLost: func(err error) RetryDecision {
				test.assert.Failf("should not be called", "err = %v", err)
				return InvalidRetryDecision
			},
			ConnectionFailed: func(attempt uint, err error) RetryDecision {
				test.Logf("Connection failed, attempt = %d, err = %v", attempt, err)
				return TryAgain
			},
			ConnectionRecovered: func() {
				test.Log("Connection is recovered after first failed attempt")
			},
		},
	})
	test.require.Error(err)

	lastPingDoneCh := make(chan struct{})
	test.healthCheckMock.On("Ping").
		Return(nil).
		Run(func(_ mock.Arguments) {
			close(lastPingDoneCh)
		}).
		Once()
	test.fakeTicker.Tick()

	<-lastPingDoneCh
	sut.Stop()

	test.assert.Equal(1+1, len(test.healthCheckMock.Calls))
}

func (test *ConnectionWatchdogTest) Test_Connection_Pings_RightAfterCreation_Initially() {
	test.healthCheckMock.On("Ping").
		Return(nil)

	sut, err := WatchConnection(test.healthCheckMock, &WatchOptions{
		NewTicker: test.returnFakeTicker,
	})
	defer sut.Stop()
	test.require.NoError(err)

	test.assert.Equal(1, len(test.healthCheckMock.Calls))
}

func (test *ConnectionWatchdogTest) Test_Connection_Retry_Recovers() {
	neverGiveUp := func(uint, error) RetryDecision {
		return TryAgain
	}

	test.healthCheckMock.On("Ping").
		Return(errors.New("broken connection")).Times(1 + 3)

	var isConnectionRecovered bool = false
	sut, err := WatchConnection(test.healthCheckMock, &WatchOptions{
		NewTicker: test.returnFakeTicker,
		On: ConnectionEvents{
			ConnectionFailed: neverGiveUp,
			ConnectionRecovered: func() {
				isConnectionRecovered = true
			},
		},
	})
	test.require.Error(err)

	test.fakeTicker.TickTimes(3)

	lastPingDoneCh := make(chan struct{})
	test.healthCheckMock.On("Ping").
		Return(nil).
		Run(func(_ mock.Arguments) {
			close(lastPingDoneCh)
		}).
		Once()
	test.fakeTicker.Tick()

	<-lastPingDoneCh
	sut.Stop() // prevent race error for `isConnectionRecovered`
	test.assert.True(isConnectionRecovered)
}

func (test *ConnectionWatchdogTest) Test_Connection_Retry_UntilGiveUp() {
	totalAttempts := 0

	tryOneMoreTime := func() OnConnectionFailedFunc {
		return func(attempt uint, err error) RetryDecision {
			totalAttempts++

			test.Logf("Expected connection attempt = %d due to error '%v'", attempt, err)
			if 3 == totalAttempts {
				return GiveUp
			}
			return TryAgain
		}
	}

	test.healthCheckMock.On("Ping").
		Return(errors.New("broken connection")).
		Times(1 + 3)

	sut, err := WatchConnection(test.healthCheckMock, &WatchOptions{
		NewTicker: test.returnFakeTicker,
		On: ConnectionEvents{
			ConnectionFailed: tryOneMoreTime(),
			ConnectionRecovered: func() {
				test.require.Fail("should not be ever called")
			},
			ConnectionLost: func(err error) RetryDecision {
				test.require.Failf("should not be ever called", "err = %v", err)
				return InvalidRetryDecision
			},
		},
	})
	test.require.Error(err)

	test.fakeTicker.TickTimes(1 + 3)

	sut.Stop()
	test.assert.Equal(3, totalAttempts)
}

func (test *ConnectionWatchdogTest) Test_InitialErrorIsNotEmpty_WhenNoConnection_OnWatchStart() {
	test.healthCheckMock.On("Ping").
		Return(errors.New("broken connection"))

	sut, err := WatchConnection(test.healthCheckMock, &WatchOptions{})
	defer sut.Stop()

	test.assert.Error(err)
}

func (test *ConnectionWatchdogTest) Test_DoesNothing_WhenStoppedAfterCreation() {
	test.healthCheckMock.On("Ping").
		Return(errors.New("broken connection"))

	sut, err := WatchConnection(test.healthCheckMock, &WatchOptions{
		NewTicker: test.returnFakeTicker,
		On: ConnectionEvents{
			ConnectionFailed: func(uint, error) RetryDecision {
				test.assert.Fail("should not be ever called")
				return InvalidRetryDecision
			},
			ConnectionRecovered: func() {
				test.assert.Fail("should not be ever called")
			},
			ConnectionLost: func(error) RetryDecision {
				test.assert.Fail("should not be ever called")
				return InvalidRetryDecision
			},
		},
	})

	sut.Stop()
	test.require.Error(err)
}
