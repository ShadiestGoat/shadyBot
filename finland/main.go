package finland

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/initializer"
	"github.com/shadiestgoat/shadyBot/utils"
)

func init() {
	initializer.Register(initializer.MOD_FINLAND, func(c *initializer.InitContext) {
		discord.MessageCreate.Add(func(s *discordgo.Session, m *discordgo.MessageCreate) bool {
			if strings.Contains(m.Content, "🇫🇮") || strings.Contains(m.Content, ":flag_fi:") || strings.Contains(strings.ToLower(m.Content), "finland") {
				go Finland(s)
			}

			return false
		})

		discord.RegisterCommand(&discordgo.ApplicationCommand{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "finland",
			Description: "send some finland love <3",
		}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
			msg, err := Finland(s)
			if log.ErrorIfErr(err, "sending finland stuff <3") {
				discutils.IError(s, i.Interaction, "Couldn't send it :(")
			}

			emb := discutils.BaseEmbed
			emb.Title = "Sent..."
			emb.Description = fmt.Sprintf(`[It's right here](%s)`, discutils.MessageURL(msg.ChannelID, msg.ID))

			discutils.IEmbed(s, i.Interaction, &emb, discutils.I_EPHEMERAL)
		})
	}, &initializer.ModuleInfo{
		ConfigOpts: []*string{&config.Channels.Finland},
		PreHooks: []initutils.Module{
			initializer.MOD_HELP_LOADER,
			initializer.MOD_DISCORD,
		},
	})
}

var components = []string{
	"LETS GOOO",
	"FINNLLAANNDND",
	"FINLAND",
	"YESSSSSS",
	":flag_fi::flag_fi:",
	":flag_fi::flag_fi::flag_fi::flag_fi::flag_fi::flag_fi::flag_fi::flag_fi:",
	"WHOOOOHOOOOOO!!!!",
	"🎉🎉",
	"FINNLLAANNDDDD",
	"MMMMMMMMMMMMMM FINLANDDD",
	"!!!!!!!!!!!!",
	"F I N L A N D",
	":flag_fi::flag_fi::flag_fi:",
	"WOOOHOOO",
	"FIIIINNNLLAAANNDDD POGGERSSSS",
	"POGGERRRSSS",
	"FINLAND FINLAND FINLAND FINLAND",
	"YESSSSSS FINLAND!!!",
}

func Finland(s *discordgo.Session) (*discordgo.Message, error) {
	content := ""

	main := utils.RandInt(0, 9)

	switch main {
	case 0:
		content = "https://tenor.com/view/astolfo-flag-finland-finland-flag-gif-25469884"
	case 1:
		content = "https://tenor.com/view/astolfo-finland-flag-anime-meme-gif-24789431"
	default:
		max := utils.RandInt(4, 9)
		for i := 0; i < max; i++ {
			comp := components[utils.RandInt(0, len(components)-1)]
			if len(comp) > 0 && comp[0] == '!' && len(content) != 0 && content[len(content)-1] == ' ' {
				content = content[:len(content)-1]
			}

			content += comp
			excl := utils.RandInt(2, 4)
			for j := 0; j < excl; j++ {
				content += "!"
			}
			content += " "
		}
		content = content[:len(content)-1]
	}

	return discutils.SendMessage(s, config.Channels.Finland, &discordgo.MessageSend{
		Content: content,
	})
}
