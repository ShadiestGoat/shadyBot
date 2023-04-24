package twitch

import (
	"time"

	"github.com/shadiestgoat/shadyBot/config"
)

var periodicMessages = []string{}

func periodicLaunch(close chan bool) {
	ticker := time.NewTicker(10 * time.Minute)
	i := 0

	for {
		select {
		case <-ticker.C:
			if len(periodicMessages) == 0 || ircClient == nil {
				continue
			}

			ircClient.Say(config.Twitch.ChannelName, periodicMessages[i])

			i++
			if i == len(periodicMessages) {
				i = 0
			}

		case <-close:
			return
		}
	}
}
