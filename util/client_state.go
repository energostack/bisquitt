package util

import (
	"fmt"
	"sync/atomic"
)

type ClientState uint32

// Client state constants.
// See MQTT-SN spec v. 1.2, chapter 6.14, p. 25
const (
	StateDisconnected ClientState = iota
	StateActive
	StateAsleep
	StateAwake
)

// String atomically get ClientState's value and converts it to human-readable string.
func (s ClientState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateActive:
		return "active"
	case StateAsleep:
		return "asleep"
	case StateAwake:
		return "awake"
	default:
		return fmt.Sprintf("unknown (%#v)", s)
	}
}

// Set atomically sets ClientState's value and returns old value.
func (s *ClientState) Set(new ClientState) ClientState {
	return ClientState(atomic.SwapUint32((*uint32)(s), uint32(new)))
}

// Get atomically gets ClientState's value.
func (s *ClientState) Get() ClientState {
	return ClientState(atomic.LoadUint32((*uint32)(s)))
}
