package connect4

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

const (
	EMB_TITLE = "🎮 - Connect 4 Invitation"
)

func init() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "connect-4",
		DefaultMemberPermissions: discord.Perms(),
		Description:       "Play connect 4 with the other player",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The other player. Note: will not ping the other player!",
				Required:    true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		target := data["user"].UserValue(nil).ID
		emb := discutils.BaseEmbed

		emb.Title = EMB_TITLE
		emb.Description = "This is an invitation to <@" + target + ">"

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: &emb,
			Comps: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Accept",
					Style:    discordgo.PrimaryButton,
					CustomID: "c4_accept_" + target,
				},
				discordgo.Button{
					Label:    "Cancel/Don't Accept",
					Style:    discordgo.DangerButton,
					CustomID: "c4_cancel_" + target,
				},
			},
		}, discutils.I_NONE)
	})

	discord.RegisterComponent("c4", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		spl := strings.Split(d.CustomID, "_")[1:]

		action := spl[0]
		targetID := spl[1]
		iAuthor := discutils.IAuthor(i.Interaction).ID

		if discutils.DefaultInteractionUpdateBtn(s, i.Interaction, targetID) {
			return
		}

		var game *Game

		switch action {
		case "cancel":
			emb := discutils.BaseEmbed
			emb.Title = EMB_TITLE
			emb.Description = "This game was either not accepted or cancelled!"
			emb.Color = discutils.COLOR_DANGER
			discutils.IEmbed(s, i.Interaction, &emb, discutils.I_UPDATE)
			return

		case "accept":
			if iAuthor != targetID {
				discutils.UpdateMessageWithSelf(s, i.Interaction)
				return
			}
			game = NewGame(discutils.IMessageAuthor(i.Interaction).ID, targetID)
		case "game":
			game = parseGame(i.Message.Embeds[0].Description)
			won := game.Action(spl[2])

			if won {
				emb := discutils.BaseEmbed
				emb.Title = EMB_TITLE
				winner := game.Player1
				if game.TurnIsPlayer2 {
					winner = game.Player2
				}
				emb.Description = "<@" + winner + "> has won! Congratulations!!"
				discutils.IEmbed(s, i.Interaction, &emb, discutils.I_UPDATE)
				return
			}
		}

		emb := discutils.BaseEmbed

		emb.Title = EMB_TITLE
		emb.Description = game.String()

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed: &emb,
			Comps: game.buttons(),
		}, discutils.I_UPDATE)
	})
}
