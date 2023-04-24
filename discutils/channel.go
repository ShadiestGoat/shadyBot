package discutils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
)

func PurgeChannel(s *discordgo.Session, chanID string) {
	before := ""
	for {
		msgs, err := s.ChannelMessages(chanID, 100, before, "", "")
		if log.ErrorIfErr(err, "Fetching message for bulk delete '%s'", chanID) {
			return
		}

		msgIDs := []string{}

		for i, msg := range msgs {
			msgIDs = append(msgIDs, msg.ID)
			if i == 99 {
				before = msg.ID
			}
		}

		if len(msgIDs) == 0 {
			break
		}

		err = s.ChannelMessagesBulkDelete(chanID, msgIDs)
		log.ErrorIfErr(err, "Bulk deleting messages on channel '%s'", chanID)
		if len(msgs) != 100 {
			break
		}
	}
}
