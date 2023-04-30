package pubsub

import (
	"math"
	"math/rand"
	"time"

	"github.com/shadiestgoat/log"
)

var failAmt = 0

var OnConnect = make(chan bool, 5)

// (re)Connect
// Should always be called as a go routine!
func Connect() {
	Close()

	if failAmt != 0 {
		d := time.Duration(math.Pow(2, float64(failAmt))) * time.Second
		d += 10 * time.Second

		if d > 2*time.Minute {
			d = 2 * time.Minute
		}

		// Small jitter <3
		d += time.Duration(math.Round(rand.Float64() * float64(time.Second)))

		time.Sleep(d)
	}

	err := start()
	if err != nil {
		failAmt++
		if failAmt > 10 {
			log.Fatal("Twitch pubsub conn error <3: %v", err)
		}
		go Connect()
		return
	} else {
		failAmt = 0
	}

	OnConnect <- true
}
