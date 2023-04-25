package xp

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/db"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func init() {
	discord.MessageCreate.Add(func(s *discordgo.Session, m *discordgo.MessageCreate) bool {
		lastMsg := time.Time{}

		err := db.QueryRowID(`SELECT last_msg FROM xp WHERE id=$1`, m.Author.ID, &lastMsg)

		if err == nil {
			if time.Since(lastMsg) >= time.Minute {
				ChangeXP(s, m.Author.ID, givenXP(config.XP.MsgMin, config.XP.MsgMax, multiplier(s, m.Author.ID)), XPS_MSG)
			}
		}

		return false
	})
}

func init() {
	announcementCloser := make(chan bool)
	vcClose := make(chan bool)

	discord.Ready.Add(func(s *discordgo.Session, v *discordgo.Ready) bool {
		go handleVCXp(s, vcClose)

		return false
	})

	discord.Ready.Add(func(s *discordgo.Session, v *discordgo.Ready) bool {
		go func() {
			for {
				select {
				case <-announcementCloser:
					return
				case e := <-XPEventChan:
					if config.Channels.XPAnnouncements == "" {
						continue
					}

					_, err := discutils.SendMessage(s, config.Channels.XPAnnouncements, &discordgo.MessageSend{
						Content: e.String(),
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{},
						},
					})
					log.ErrorIfErr(err, "sending xp announcement")
				}
			}
		}()

		return false
	})

	initializer.RegisterCloser(initializer.MOD_XP, func() {
		announcementCloser <- true
		vcClose <- true
	})
}
