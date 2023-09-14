package pubsub

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/shadiestgoat/log"
	"nhooyr.io/websocket"
)

var msg_ping []byte

func init() {
	msg_ping, _ = json.Marshal(resp{
		Type: "PING",
	})
}

var pongDone = atomic.Bool{}
var closePing = make(chan bool, 2)

func startPing(origin string) {
	if doingPing.Load() {
		log.Warn("Doing ping x2 (Origin: %s)", origin)
		return
	}
	t := time.NewTicker(4 * time.Minute)

	doingPing.Store(true)

	defer func() {
		doingPing.Store(false)
	}()

	for {
		select {
		case <-t.C:
			doPing()
		case <-closePing:
			return
		}
	}
}

func doPing() {
	if wsConn == nil {
		log.Warn("PubSub WS Conn is nil when pinging!")
		return
	}

	err := wsConn.Write(context.Background(), websocket.MessageText, msg_ping)
	if log.ErrorIfErr(err, "writing to pubsub conn") {
		go Connect("doPing: post error")
		return
	}
	time.Sleep(10 * time.Second)

	if !pongDone.Load() {
		log.Error("Pong not received :(")
		Close("Ping Fail")
		go Connect("doPing: post non-pong")
		return
	}

	pongDone.Store(false)
}
