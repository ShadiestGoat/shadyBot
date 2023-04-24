package misc

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

// Prefix chance:  0.25
// Suffix chance:  0.25

var vowels = map[rune]bool{
	'e': true,
	'y': true,
	'u': true,
	'i': true,
	'o': true,
	'a': true,
}

var wordsReplace = map[string][]string{
	"love":   {"wuv"},
	"mr":     {"mista", "mistuh", "sir"},
	"cat":    {"catto", "kitten", "kitteh"},
	"hello":  {"hiii", "henwo", "hewwo"},
	"hell":   {"heck", "hecc"},
	"thick":  {"thicc"},
	"friend": {"fwend"},
	"stop":   {"stawp"},
}

var suffixes = []string{
	"(ﾉ´ з `)ノ",
	"( ´ ▽ ` ).｡ｏ♡",
	"(´,,•ω•,,)♡",
	"(*≧▽≦)",
	"ɾ⚈▿⚈ɹ",
	"( ﾟ∀ ﾟ)",
	"( ・ ̫・)",
	"( •́ .̫ •̀ )",
	"(▰˘v˘▰)",
	"(・ω・)",
	"✾(〜 ☌ω☌)〜✾",
	"(ᗒᗨᗕ)",
	"(・`ω´・)",
	":3",
	">:3",
	"hehe",
	"xox",
	">3<",
	"murr~",
	"UwU",
	"nya~",
}

func init() {
	for w := range wordsReplace {
		ogW := w
		w = owoLetters(w)

		if w == ogW {
			continue
		}

		wordsReplace[w] = wordsReplace[ogW]
		delete(wordsReplace, ogW)
	}
}

func owoLetters(inp string) string {
	return strings.TrimSpace(strings.NewReplacer("r", "w", "l", "w").Replace(inp))
}

func MakeOwO(inp string) string {
	inp = owoLetters(inp)

	words := strings.Split(inp, " ")
	newWords := ""

	for _, w := range words {
		if len(w) == 0 {
			continue
		}

		// stutter
		if utils.Chance(0.05) {
			w = w[:1] + "-" + w[1:]
		}

		if w[len(w)-1] == 'y' {
			vowelLoc := -1

			for i, r := range w {
				if vowels[r] {
					vowelLoc = i
					break
				}
			}

			if vowelLoc != -1 {
				w += "-w" + w[vowelLoc:]
			}
		}

		if vowels[rune(w[len(w)-1])] {
			c := 0.25

			r := rune(w[len(w)-1])

			for {
				if !utils.Chance(c) {
					break
				}
				w += string(r)

				if utils.Chance(c + 0.5) {
					w += string(r)
				}
				if utils.Chance(c + 0.1) {
					w += string(r)
				}

				c -= 0.05
			}
		}

		newWords += " " + w
	}

	if len(newWords) != 0 {
		newWords = newWords[1:]
	}

	if utils.Chance(0.25) {
		newWords += " " + suffixes[utils.RandInt(0, len(suffixes)-1)]
	}

	return newWords
}

func cmdOwo() {
	discord.RegisterCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "owoify",
		Description: "owo-ify some text ~",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "the text you want to owo-ify",
				Required:    true,
			},
		},
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
		text := data["text"].StringValue()
		out := MakeOwO(text)

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Content: &out,
		}, discutils.I_EPHEMERAL)
	})
}
