package websocket

import (
	"sync/atomic"
)

type ServiceState int64

const (
	// StoppedState service stopped
	StoppedState ServiceState = iota + 1
	// StartingState service starting
	StartingState
	// RunningState service running
	RunningState
	// StoppingState service stopping
	StoppingState
)

func (s *ServiceState) Store(state ServiceState) {
	atomic.StoreInt64((*int64)(s), int64(state))
}

func (s *ServiceState) Load() ServiceState {
	return ServiceState(atomic.LoadInt64((*int64)(s)))
}
