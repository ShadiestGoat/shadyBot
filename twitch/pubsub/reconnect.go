package pubsub

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/shadiestgoat/log"
)

var failAmt = 0

// Fires when a successful connection is done. Returns the connection origin
var OnConnect = make(chan string)

var connectionLock = atomic.Bool{}

// (re)Connect
// Should always be called as a goroutine!
func Connect(connOrigin string) {
	if connectionLock.Load() {
		return
	}
	connectionLock.Store(true)

	Close()

	// Always sleep, even if its for like 10 seconds
	d := time.Duration(math.Pow(2, float64(failAmt))) * time.Second
	d += 10 * time.Second

	if d > 2*time.Minute {
		d = 2 * time.Minute
	}

	// Small jitter <3
	d += time.Duration(math.Round(rand.Float64() * float64(time.Second)))

	time.Sleep(d)

	err := start("connOrigin")
	if err != nil {
		failAmt++
		if failAmt > 10 {
			log.Fatal("Twitch pubsub conn error <3: %v", err)
		}

		go Connect(connOrigin)
		return
	} else {
		failAmt = 0
	}

	OnConnect <- connOrigin

	connectionLock.Store(false)
}
