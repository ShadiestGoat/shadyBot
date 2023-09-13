package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var wsConn *websocket.Conn
var authToken string

func SetToken(t string) {
	authToken = t
}

var errDoubleConn = errors.New("double conn")

type stupidErr struct {
	ctx string
	err error
}

func (err stupidErr) Error() string {
	return err.ctx + ": " + err.err.Error()
}

type respErr struct {
	msg string
}

func (err respErr) Error() string {
	return err.msg
}

// Returns if the connection is successful or not
func start(connOrigin string) error {
	if wsConn != nil {
		log.Warn("Pubsub open x2!! (Origin: %s)", connOrigin)
		return errDoubleConn
	}

	conn, _, err := websocket.Dial(context.Background(), "wss://pubsub-edge.twitch.tv", nil)

	if err != nil {
		return &stupidErr{
			ctx: "Dialing twitch pubsub",
			err: err,
		}
	}

	wsConn = conn

	body, _ := json.Marshal(listen{
		Topics: []string{"channel-points-channel-v1." + config.Twitch.OwnID},
		Auth:   authToken,
	})

	wsjson.Write(context.Background(), wsConn, resp{
		Type: "LISTEN",
		Data: body,
	})

	out := &resp{}

	err = wsjson.Read(context.Background(), wsConn, out)

	if err != nil {
		return &stupidErr{
			ctx: "reading listen resp for pubsub",
			err: err,
		}
	}

	if out.Error != "" {
		return &stupidErr{
			ctx: "pubsub resp",
			err: &respErr{
				msg: out.Error,
			},
		}
	}

	isClosing.Store(false)
	go startPing(connOrigin)
	go startReading(connOrigin)

	return nil
}

var Redeems = make(chan *Redemption, 10)

func startReading(origin string) {
	if doingRead.Load() {
		log.Warn("Reading x2 !!! (Origin: %s)", origin)
		return
	}

	doingRead.Store(true)

	defer func() {
		doingRead.Store(false)
	}()

	for {
		_, msg, err := wsConn.Reader(context.Background())
		if isClosing.Load() {
			go Connect("startReading: isClosing load")
			return
		}

		if err != nil {
			if !errors.As(err, &websocket.CloseError{}) {
				log.Error("While reading twitch pubsub: %v", err)
				return
			}

			go Connect("startReading: error")
			return
		}

		v := &resp{}
		log.ErrorIfErr(json.NewDecoder(msg).Decode(&v), "decoding pubsub")

		switch v.Type {
		case "PONG":
			pongDone.Store(true)
		case "reward-redeemed":
			rawR := rawReward{}
			json.Unmarshal(v.Data, &rawR)
			Redeems <- rawR.Parse()
		case "RECONNECT":
			go Connect("startReading: RECONNECT")
			return
		}
	}
}

var isClosing = atomic.Bool{}

func Close() {
	if isClosing.Load() || wsConn == nil {
		return
	}

	closePing <- true
	wsConn.Close(websocket.StatusGoingAway, "Cya <3")
	isClosing.Store(true)
}

type resp struct {
	Type  string          `json:"type"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type listen struct {
	Topics []string `json:"topics"`
	Auth   string   `json:"auth_token"`
}
