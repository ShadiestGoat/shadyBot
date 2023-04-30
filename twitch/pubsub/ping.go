package pubsub

import (
	"context"
	"encoding/json"
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

var pongDone = false
var closePing = make(chan bool, 2)

func startPing() {
	if doingPing {
		log.Warn("Doing ping x2!!!!")
		return
	}
	t := time.NewTicker(4 * time.Minute)

	doingPing = true

	defer func() {
		doingPing = false
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
		go Connect()
		return
	}
	time.Sleep(10 * time.Second)
	if !pongDone {
		log.Error("Pong not received :(")
		go Connect()
		return
	}
	pongDone = false
}
