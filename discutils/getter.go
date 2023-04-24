package discutils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
)

func GetChannel(s *discordgo.Session, channelID string) *discordgo.Channel {
	c, err := s.State.Channel(channelID)
	if err == nil && c != nil {
		return c
	}

	c, err = s.Channel(channelID)

	if err != nil {
		log.ErrorIfErr(err, "fetching channel '%s'", channelID)
		c = nil
	}

	return c
}

func GetMember(s *discordgo.Session, guildID, userID string) *discordgo.Member {
	m, err := s.State.Member(guildID, userID)
	if err == nil && m != nil {
		return m
	}
	m, err = s.GuildMember(guildID, userID)
	if err != nil {
		log.ErrorIfErr(err, "fetching member '%s' in guild '%s'", userID, guildID)
		m = nil
	}

	return m
}

func GetMessage(s *discordgo.Session, channelID, msgID string) *discordgo.Message {
	m, err := s.State.Message(channelID, msgID)
	if err == nil && m != nil {
		return m
	}
	m, err = s.ChannelMessage(channelID, msgID)
	if err != nil {
		log.ErrorIfErr(err, "fetching message '%s' in channel '%s'", msgID, channelID)
		m = nil
	}

	return m
}
