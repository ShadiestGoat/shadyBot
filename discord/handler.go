package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
)

type HandlerFunc[T any] func(s *discordgo.Session, v T) bool

// Handle an event, and return if the event has been fully handled.
// If it returns true, no subsequent handlers will be called.
type Handler[T any] []HandlerFunc[T]

func (allHandlers Handler[T]) Handle(s *discordgo.Session, v T) {
	for _, h := range allHandlers {
		r := h(s, v)
		if r {
			break
		}
	}
}

// Handle an event, and return true if the event has been fully handled.
// If it returns true, no subsequent handlers will be called.
func (allHandlers *Handler[T]) Add(f HandlerFunc[T]) {
	*allHandlers = append(*allHandlers, f)
}

var MessageReactionAdd = Handler[*discordgo.MessageReactionAdd]{}

var MemberJoin = Handler[*discordgo.GuildMemberAdd]{}

var MessageReactionRemove = Handler[*discordgo.MessageReactionRemove]{}

var MessageCreate = Handler[*discordgo.MessageCreate]{
	func(s *discordgo.Session, m *discordgo.MessageCreate) bool {
		if m.GuildID != config.Discord.GuildID || m.Author.Bot {
			return true
		}

		return false
	},
}

var MessageRemove = Handler[*discordgo.MessageDelete]{}

var MessageUpdate = Handler[*discordgo.MessageUpdate]{}

var Ready = Handler[*discordgo.Ready]{}
