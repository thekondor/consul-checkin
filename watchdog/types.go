package watchdog

type RetryDecision uint8

const (
	GiveUp RetryDecision = iota
	TryAgain
)

func (self RetryDecision) bool() bool {
	switch self {
	case TryAgain:
		return true
	case GiveUp:
		return false
	default:
		panic("Invalid retry decision provided")
	}
}

type (
	OnConnectionFailedFunc    func(attempt uint, err error) RetryDecision
	OnConnectionLostFunc      func(err error) RetryDecision
	OnConnectionRecoveredFunc func()
)
