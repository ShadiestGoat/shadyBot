package discutils

import "github.com/bwmarrin/discordgo"

// Generates a string with all the message attachment urls on a newline.
func AttachmentsString(msg *discordgo.Message) string {
	if msg == nil || len(msg.Attachments) == 0 {
		return ""
	}

	a := ""

	for _, att := range msg.Attachments {
		a += att.URL + "\n"
	}

	return a[:len(a)-1]
}
