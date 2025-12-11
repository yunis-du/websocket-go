package websocket

import "sync"

// RWMutex is a wrapper around sync.RWMutex
type RWMutex struct {
	sync.RWMutex
}
