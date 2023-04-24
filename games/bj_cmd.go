package games

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/xp"
)

func cmdBlackjack() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "blackjack",
		Description: "Bet xp on a blackjack game",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "bet",
				Description: "The XP you are betting",
				Required:    true,
				MinValue:    &config.BET_MIN,
				MaxValue:    config.BET_MAX,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		bet := int(data["bet"].IntValue())
		if InActivityErrorCheck(s, i.Interaction) {
			return
		}
		userID := discutils.IAuthor(i.Interaction).ID
		curXP, err := xp.FetchXP(userID)
		if err != nil {
			discutils.IError(s, i.Interaction, "Couldn't fetch your XP!")
			return
		}

		if !curXP.HasNeededXP(bet) {
			discutils.IError(s, i.Interaction, "Your bet was too large! You can't afford it (L)")
			return
		}

		game := &BJGame{
			UserID:     userID,
			Bet:        bet,
			Doubled:    false,
			UserHand:   []int{},
			DealerHand: []int{},
			UserTurn:   true,
		}

		game.UserHand = append(game.UserHand, game.NewCard())
		game.DealerHand = append(game.DealerHand, game.NewCard())

		content := "<@" + userID + ">"

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed:   game.embedGame(),
			Comps:   game.buttons(),
			Content: &content,
		}, discutils.I_NONE)

		ActivityStore.Add(i.Interaction, GT_BJ)
	})

	discord.RegisterComponent("bj", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		action := strings.Split(d.CustomID, "_")[1]

		var game *BJGame

		if len(i.Message.Embeds) == 1 {
			game = bjParseGame(i.Message.Embeds[0])
		}

		if game == nil {
			discutils.IError(s, i.Interaction, "Couldn't parse the game :(", discutils.I_UPDATE)
		}

		if discutils.DefaultInteractionUpdateBtn(s, i.Interaction) {
			return
		}

		ActivityStore.Update(i.Interaction)

		userBusted := false

		responded := discutils.I_UPDATE

		switch action {
		case "double":
			curXP, err := xp.FetchXP(game.UserID)
			shouldWait := true

			if err != nil {
				discutils.IErrorBtn(s, i.Interaction, "Couldn't fetch your xp :(", false)
			} else if !curXP.HasNeededXP(game.Bet * 2) {
				discutils.IErrorBtn(s, i.Interaction, "Couldn't fetch your xp :(", false)
			} else {
				shouldWait = false
			}

			if shouldWait {
				responded = discutils.I_EDIT

				time.Sleep(1500 * time.Millisecond)
				btn := game.buttons()

				// IErrorBtn responded
				discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
					Embed: game.embedGame(),
					Comps: btn,
				}, responded)

				return
			}

			game.Doubled = true
			game.UserTurn = false

			fallthrough
		case "hit":
			game.UserHand = append(game.UserHand, game.NewCard())
			if len(game.UserHand.Totals(true)) == 0 {
				userBusted = true
			}
		case "stand":
			game.UserTurn = false
		}

		msgToSend := ""
		won := false
		gotDraw := false

		if userBusted {
			won = false
			msgToSend = "You busted"
		} else if game.UserHand.Totals(true)[0] == 21 {
			won = true
			msgToSend = "**You** got the *femboy BJ*"
		} else if !game.UserTurn {
			// Dealer's turn

			discutils.IResp(s, i.Interaction, game.respOpts(), responded)
			
			responded = discutils.I_EDIT

			var state = game.dealerLoopState()

			for state == bjd_continue {
				time.Sleep(800 * time.Millisecond)
				game.DealerHand = append(game.DealerHand, game.NewCard())

				discutils.IResp(s, i.Interaction, game.respOpts(), responded)
				state = game.dealerLoopState()
			}

			switch state {
			case bjd_bust:
				won = true
				msgToSend = "The Dealer busted"
			case bjd_draw:
				gotDraw = true
			case bjd_lost:
				won = true
				msgToSend = "Your hand is greater than the dealer's"
			}
		}

		if gotDraw || msgToSend != "" {
			FinishGame(game.UserID, 0, false, GT_BJ)
			emb := discutils.BaseEmbed
			emb.Title = TITLE_BJ
			emb.Description = "The dealer's hand is **equal** to yours! This means you **neither won nor lost** any xp." + game.handString()

			discutils.IEmbed(s, i.Interaction, &emb, responded)
		} else if msgToSend != "" {
			FinishGame(game.UserID, game.TrueBet(), won, GT_BJ)
			discutils.IEmbed(s, i.Interaction, game.embedBase(msgToSend, won), responded)
		} else {
			discutils.IResp(s, i.Interaction, game.respOpts(), responded)
		}
	})
}

func (g BJGame) respOpts() *discutils.IRespOpts {
	return &discutils.IRespOpts{
		Embed:   g.embedGame(),
		Comps:   g.buttons(),
		Content: g.msgContent(),
	}
}

func (g BJGame) buttons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.Button{
			Disabled: !g.UserTurn,
			Style:    discordgo.PrimaryButton,
			Label:    "Hit",
			CustomID: "bj_hit",
		},
		discordgo.Button{
			Disabled: !g.UserTurn,
			Style:    discordgo.DangerButton,
			Label:    "Stand",
			CustomID: "bj_stand",
		},
		discordgo.Button{
			Disabled: !g.UserTurn,
			Style:    discordgo.PrimaryButton,
			Label:    "Double Down",
			CustomID: "bj_double",
		},
	}
}

func (g BJGame) msgContent() *string {
	str := "<@" + g.UserID + ">"
	return &str
}
