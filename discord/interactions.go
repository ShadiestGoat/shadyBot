package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discutils"
)

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		d := i.ApplicationCommandData()
		if h, ok := commandHandlers[d.Name]; ok {
			h(s, i, &d, discutils.ParseCommand(d))
		}
	case discordgo.InteractionModalSubmit:
		d := i.ModalSubmitData()
		if h, ok := modalHandlers[strings.SplitN(d.CustomID, "_", 2)[0]]; ok {
			h(s, i, &d, discutils.ParseModal(d))
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		d := i.ApplicationCommandData()
		if h, ok := autocompleteHandlers[d.Name]; ok {
			h(s, i, &d, discutils.ParseCommand(d))
		}
	case discordgo.InteractionMessageComponent:
		d := i.MessageComponentData()
		if h, ok := componentHandlers[strings.SplitN(d.CustomID, "_", 2)[0]]; ok {
			h(s, i, &d)
		}
	}
}

type HandlerCommand func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption)
type HandlerModal func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ModalSubmitInteractionData, data map[string]string)
type HandlerComponent func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData)

var commandHandlers = map[string]HandlerCommand{}
var autocompleteHandlers = map[string]HandlerCommand{}
var modalHandlers = map[string]HandlerModal{}
var componentHandlers = map[string]HandlerComponent{}

func RegisterAutocomplete(cmd string, handler HandlerCommand) {
	autocompleteHandlers[cmd] = handler
}

func RegisterModal(id string, handler HandlerModal) {
	modalHandlers[id] = handler
}

func RegisterComponent(id string, handler HandlerComponent) {
	componentHandlers[id] = handler
}
