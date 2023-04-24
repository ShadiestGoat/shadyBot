package discord

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func printCmd(cmd *discordgo.ApplicationCommand) {
	b1, _ := json.Marshal(cmd)

	fmt.Println(string(b1))
}

func commandIsEqual(cmd1, cmd2 *discordgo.ApplicationCommand) bool {
	if cmd1.DefaultPermission == nil {
		cmd1.DefaultPermission = DefaultTrue
	}
	if cmd2.DefaultPermission == nil {
		cmd2.DefaultPermission = DefaultTrue
	}
	if cmd1.DefaultMemberPermissions == nil {
		cmd1.DefaultMemberPermissions = Perms()
	}
	if cmd2.DefaultMemberPermissions == nil {
		cmd2.DefaultMemberPermissions = Perms()
	}

	if cmd1.Description != cmd2.Description || len(cmd1.Options) != len(cmd2.Options) {
		log.Debug("C 1")
		return false
	}

	if *cmd1.DefaultPermission != *cmd2.DefaultPermission || *cmd1.DefaultMemberPermissions != *cmd2.DefaultMemberPermissions {
		log.Debug("Perm stuff")
		printCmd(cmd1)
		printCmd(cmd2)
		return false
	}

	for i := range cmd1.Options {
		if !optIsEqual(cmd1.Options[i], cmd2.Options[i]) {
			log.Debug("Opt issue index %d", i)
			printCmd(cmd1)
			printCmd(cmd2)

			return false
		}
	}

	return true
}

func optIsEqual(opt1, opt2 *discordgo.ApplicationCommandOption) bool {
	if opt1.Name != opt2.Name || opt1.Description != opt2.Description || opt1.Type != opt2.Type || opt1.Required != opt2.Required {
		log.Debug("Name issue")
		return false
	}

	if len(opt1.Options) != len(opt2.Options) || len(opt1.Choices) != len(opt2.Choices) || len(opt1.ChannelTypes) != len(opt2.ChannelTypes) {
		log.Debug("Len issue")
		return false
	}

	for i := range opt1.ChannelTypes {
		if opt1.ChannelTypes[i] != opt2.ChannelTypes[i] {
			log.Debug("CT issue")
			return false
		}
	}

	for i := range opt1.Choices {
		c1 := opt1.Choices[i]
		c2 := opt2.Choices[i]

		if c1.Name != c2.Name || fmt.Sprint(c1.Value) != fmt.Sprint(c2.Value) {
			return false
		}
	}

	for i := range opt1.Options {
		if !optIsEqual(opt1.Options[i], opt2.Options[i]) {
			return false
		}
	}

	return true
}

// TODO: Modularize the intents
func init() {
	var s *discordgo.Session

	RegisterComponent("cancel", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		emb := discutils.BaseEmbed

		emb.Title = "Cancelled"
		emb.Description = "This operation was cancelled!"
		emb.Color = discutils.COLOR_DANGER

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed:   &discordgo.MessageEmbed{},
			Comps:   []discordgo.MessageComponent{},
			Content: new(string),
		}, discutils.I_UPDATE)
	})

	RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "bugs",
		Description: "View the URL to report bugs",
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		emb := discutils.BaseEmbed
		emb.Title = "Bugs"
		bugURL := "https://github.com/shadiestgoat/shadyBot/issues"
		emb.Description = "[You can report bugs here](" + bugURL + ")"
		emb.URL = bugURL
		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_EPHEMERAL)
	})

	initializer.Register(initializer.MOD_DISCORD, func(c *initializer.InitContext) {
		var err error
		s, err = discordgo.New("Bot " + config.Discord.Token)
		log.FatalIfErr(err, "Creating a discord session")

		s.Identify.Intents = discordgo.IntentGuildPresences |
			discordgo.IntentGuilds |
			discordgo.IntentGuildVoiceStates |
			discordgo.IntentGuildMessages |
			discordgo.IntentGuildPresences |
			discordgo.IntentMessageContent |
			discordgo.IntentGuildMembers

		s.AddHandler(handleInteraction)
		s.AddHandler(MessageReactionAdd.Handle)
		s.AddHandler(MemberJoin.Handle)
		s.AddHandler(MessageReactionRemove.Handle)
		s.AddHandler(MessageCreate.Handle)
		s.AddHandler(MessageRemove.Handle)

		Ready.Add(func(s *discordgo.Session, e *discordgo.Ready) bool {
			log.Success("Discord connected as '%s'", e.User.Username)
			return false
		})

		s.AddHandlerOnce(Ready.Handle)

		s.StateEnabled = true
		s.State.MaxMessageCount = 150

		err = s.Open()
		log.FatalIfErr(err, "Opening the Discord connection")

		appID := s.State.User.ID

		curCommands, err := s.ApplicationCommands(appID, "")
		log.FatalIfErr(err, "fetching application commands")

		oldCommands := map[string]*discordgo.ApplicationCommand{}

		for i, cmd := range curCommands {
			oldCommands[cmd.Name] = curCommands[i]

			if _, ok := commands[cmd.Name]; !ok {
				err := s.ApplicationCommandDelete(appID, "", cmd.ID)
				if log.ErrorIfErr(err, "deleting command '%s' (id: '%s')", cmd.Name, cmd.ID) {
					log.Warn("This means that the command '%s' will still be there, but it has no handler!", cmd.Name)
				} else {
					log.Debug("Removed command '%s' as non existent", cmd.Name)
				}
			}
		}

		log.Debug("Deprecated commands removed")

		for _, v := range commands {
			if oldCommands[v.Name] == nil || !commandIsEqual(oldCommands[v.Name], v) {
				v, err = s.ApplicationCommandCreate(s.State.User.ID, "", v)
				log.FatalIfErr(err, "creating/updating command '%s' (id: '%s')", v.Name, v.ID)
				log.Success("Uploaded command '%s'", v.Name)
			} else {
				v = oldCommands[v.Name]
				log.Debug("Skipping adding command '%s'", v.Name)
			}

			// Update with an ID, etc
			commands[v.Name] = v
		}

		c.Discord = s
	}, nil)

	initializer.RegisterCloser(initializer.MOD_DISCORD, func() {
		s.Close()
	})
}
