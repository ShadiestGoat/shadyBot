package pronouns

import (
	"fmt"

	"github.com/ShadiestGoat/pronoundb"
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
)

var client = pronoundb.NewClient()

func Discord(userID string) (pronoundb.Pronoun, error) {
	return client.Lookup(pronoundb.PLATFORM_DISCORD, userID)
}

func Twitch(userID string) (pronoundb.Pronoun, error) {
	return client.Lookup(pronoundb.PLATFORM_TWITCH, userID)
}

func Explain(pr pronoundb.Pronoun) string {
	switch pr {
	case pronoundb.PR_UNSPECIFIED:
		return "This user's pronouns are unspecified. This usually means that they have yet to add themselves onto pronoundb. It is usually safe to use they/them pronouns in this case!"
	case pronoundb.PR_ANY:
		return "This user stated that any pronouns are fine. Go wild!"
	case pronoundb.PR_ASK:
		return "This user prefers that you ask for the correct pronouns!"
	case pronoundb.PR_AVOID:
		return "This user indicated that you should avoid the use of any pronouns!"
	case pronoundb.PR_OTHER:
		return "This user said that the correct pronouns are not yet listed on pronoundb, in which case, please ask for the correct pronouns!"
	}

	genders := pr.Genders()
	switch len(genders) {
	case 1:
		return fmt.Sprintf("This user specified that %s pronouns are **%s**, therefore, please refer %s as such!", pr.Their(), pr.Abbreviation(), pr.Them())
	case 2:
		return fmt.Sprintf("This user specified that %s pronouns are **%s**. This usually means that %s are ok with *%s* pronouns, but prefer **%s**!", pr.Their(), pr.Abbreviation(), pr.They(), genders[1].Abbreviation(), genders[0].Abbreviation())
	}

	return ""
}

func init() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "pronouns",
		DefaultMemberPermissions: discord.Perms(),
		Description:              "Get information about a user's pronouns",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to get pronouns of",
				Required:    false,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		var user *discordgo.User

		if data["user"] != nil {
			user = data["user"].UserValue(s)
		} else {
			user = discutils.IAuthor(i.Interaction)
		}

		pr, err := Discord(user.ID)
		if log.ErrorIfErr(err, "fetching discord pronouns for user '%s'", user.ID) {
			discutils.IError(s, i.Interaction, "Couldn't fetch pronouns :(")
			return
		}

		emb := discutils.BaseEmbed
		emb.Title = "Pronouns"
		emb.Description = "**" + pr.Abbreviation() + "**\n" + Explain(pr) + "\n\n[Pronouns are sourced from pronoundb.org](https://pronoundb.org)"

		if user != nil {
			emb.Author = &discordgo.MessageEmbedAuthor{
				IconURL: user.AvatarURL("256"),
				Name:    user.Username,
			}
		}

		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_NONE)
	})
}
