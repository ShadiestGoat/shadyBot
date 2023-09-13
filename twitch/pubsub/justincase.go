package pubsub

import "sync/atomic"

var (
	doingPing = atomic.Bool{}
	doingRead = atomic.Bool{}
)
