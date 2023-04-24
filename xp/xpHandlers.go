package xp

import (
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func handleVCXp(s *discordgo.Session, closer chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-closer:
			return
		case <-ticker.C:
			go func() {
				g, _ := s.State.Guild(config.Discord.GuildID)

				chanCount := map[string]int{}

				for _, vc := range g.VoiceStates {
					chanCount[vc.ChannelID]++
				}

				for _, vc := range g.VoiceStates {
					if vc.Deaf || vc.SelfDeaf || vc.Mute {
						continue
					}
					mem := vc.Member
					if mem == nil {
						mem = discutils.GetMember(s, config.Discord.GuildID, vc.UserID)
						if mem == nil {
							log.Warn(`User '%v' ignored for vc state`, vc.UserID)
							continue
						}
					}

					multiplier := 1.0

					if vc.SelfStream && vc.SelfVideo {
						multiplier *= config.XP_VC_VIDEO_AND_STREAM
					} else if vc.SelfStream || vc.SelfVideo {
						multiplier *= config.XP_VC_VIDEO
					}

					if vc.SelfMute {
						multiplier *= config.XP_VC_MUTE
					}

					if chanCount[vc.ChannelID] < 2 {
						multiplier *= config.XP_VC_ALONE
					}

					xpGiven := int(math.Round(float64(utils.RandInt(config.XP_VC_MIN, config.XP_VC_MAX)) * multiplier))

					ChangeXP(s, vc.UserID, xpGiven, XPS_VC)
				}
			}()
		}
	}
}
