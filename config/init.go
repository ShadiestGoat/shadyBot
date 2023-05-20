package config

import (
	"errors"
	"os"
	"strings"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func init() {
	initializer.RegisterPriority(initializer.MOD_CONFIG, func(c *initializer.InitContext) {
		load()

		_, err := os.Stat("resources")
		if err != nil {
			os.Mkdir("resources", 0755)
			os.Mkdir("resources/twitch", 0755)
		}

		b, err := os.ReadFile("resources/twitch/so.md")
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Warn("Can't open the custom shoutout file, twitch !so command will have no custom names :(")
			} else {
				log.Warn("There are no custom shoutouts!")
			}
		} else {
			lines := strings.Split(string(b), "\n")
			for _, l := range lines {
				info := strings.SplitN(l, " - ", 2)
				if len(info) != 2 {
					log.Warn("'%s' is not a valid so line. Format must be 'name - custom message'", l)
					continue
				}
				info[0] = strings.ToLower(info[0])
				Twitch.CustomSO[info[0]] = info[1]
			}
		}

		Twitch.ChannelName = strings.ToLower(Twitch.ChannelName)
		
		c.DisabledModules = General.Disabled
	}, nil)

	initializer.RegisterPriority(initializer.MOD_LOG, func(c *initializer.InitContext) {
		initLog()
	}, nil, initializer.MOD_CONFIG)

	initializer.RegisterCloser(initializer.MOD_LOG, log.Close)
}

func initLog() {
	logCBs := []log.LogCB{
		log.NewLoggerPrint(),
		log.NewLoggerFileComplex("logs/log", log.FILE_DESCENDING, 5),
	}

	if debugV.WebHook != "" {
		if debugV.Mention != "" {
			debugV.Mention += ", "
		}

		logCBs = append(logCBs, log.NewLoggerDiscordWebhook(debugV.Mention, debugV.WebHook))
	}

	log.Init(logCBs...)
}
