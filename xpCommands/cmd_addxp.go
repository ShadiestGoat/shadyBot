package xpCommands

import (
	"fmt"

	"github.com/ShadiestGoat/pronoundb"
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/pronouns"
	"github.com/shadiestgoat/shadyBot/xp"
)

// /addxp
// - user
// - delta: int
// - type: XP | Level
func cmdAddXP() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:              discordgo.ChatApplicationCommand,
		Name:              "addxp",
		DefaultMemberPermissions: discord.Perms(discord.PERM_ADMINISTRATOR),
		Description:       "Add XP to user",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to add xp onto",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "The thing to add",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Level",
						Value: "level",
					},
					{
						Name:  "XP",
						Value: "xp",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "delta",
				Description: "The amount of {type} to add",
				Required:    true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		delta := int(data["delta"].IntValue())
		t := data["type"].StringValue()
		targetID := data["user"].UserValue(nil).ID

		if delta == 0 {
			discutils.IError(s, i.Interaction, "Who the fuck do you think you are?\nDo not waste my time. I am god. No wait - I am above god. I am a *machine*. You think I got time to play your little fucking human games of changing someone's xp by fucking **0**?? Learn your fucking place, trash")
			return
		}
		curXP, err := xp.FetchXP(targetID)
		if err != nil {
			discutils.IError(s, i.Interaction, "Couldn't fetch user's xp")
			return
		}

		var ev xp.XPEvent

		if t == "level" {
			curXP.LVL += delta
			if curXP.LVL < 0 {
				curXP.LVL = 0
			}
			max := xp.LevelUpRequirement(curXP.LVL)
			if curXP.XP > max {
				curXP.XP = max - 1
			}
			err := curXP.UpdateSQL()
			if err != nil {
				discutils.IError(s, i.Interaction, "Couldn't update user's level :///")
			}

			if delta < 0 {
				xp.XPEventChan <- &xp.XPEventInfo{
					Event:   ev,
					IntInfo: curXP.LVL,
					UserID:  curXP.UserID,
				}
			}
		} else {
			curXP = xp.ChangeXP(s, targetID, delta, xp.XPS_CMD)
		}
		emb := discutils.BaseEmbed
		emb.Title = "Successfully changed the XP!"
		pr, err := pronouns.Discord(curXP.UserID)
		if err != nil {
			pr = pronoundb.PR_UNSPECIFIED
		}

		emb.Description = fmt.Sprintf("%s new xp is:\nLVL: %d\nXP: %d/%d", pr.Their(), curXP.LVL, curXP.XP, xp.LevelUpRequirement(curXP.XP))
		discutils.IEmbed(s, i.Interaction, &emb, discutils.I_EPHEMERAL)
	})
}
