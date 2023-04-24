package help

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/initializer"
)

var htmlCommentsReg = regexp.MustCompile(`<!--.+?-->`)
var mdLineSep = regexp.MustCompile(`(?m)^[^\S\n]*-+?[^\S\n]*$`)

type sectionStage1 struct {
	Name    string
	Content string
	Size    int
}

type sectionStage2 struct {
	Name    string
	Content string
	Size    int

	Parent   *sectionStage2 `json:"-"`
	Children []*sectionStage2
}

func (cmd sectionStage2) CommandString(cmdName string, subCommands string) string {
	mention := discord.CommandMention(cmdName, subCommands)
	if mention == "" {
		return ""
	}

	str := ""

	if cmd.Content != "" {
		str = mention + " - " + cmd.Content
	} else if cmd.Size == 2 {
		str = mention
		if len(cmd.Children) != 0 {
			str += ":"
		}
	}

	if len(cmd.Children) != 0 {
		str += "\n"

		sub := subCommands

		if sub != "" {
			sub += " " + cmd.Name
		}

		for _, c := range cmd.Children {
			str += "◈ " + c.CommandString(cmdName, sub) + "\n"
		}

		str = str[:len(str)-1]
	}

	return str
}

func init() {
	initializer.Register(initializer.MOD_HELP_LOADER, func(c *initializer.InitContext) {
		// We already know this will not be an error
		b, _ := os.ReadFile("resources/help.md")
		b = mdLineSep.ReplaceAll(htmlCommentsReg.ReplaceAll(b, []byte{}), []byte{})

		lines := strings.Split(string(b), "\n")

		tmpAllSections := []*sectionStage1{}

		curSec := &sectionStage1{
			Name:    "Root",
			Content: "",
			Size:    0,
		}

		for _, l := range lines {
			if len(l) == 0 {
				continue
			}

			size := 0

			for i, r := range l {
				if r != '#' {
					if r == ' ' && i+1 < len(l) {
						size = i
					}

					break
				}
			}

			if size == 0 {
				curSec.Content += "\n" + l
			} else {
				tmpAllSections = append(tmpAllSections, curSec)

				curSec = &sectionStage1{
					Name:    l[size+1:],
					Content: "",
					Size:    size,
				}
			}
		}

		tmpAllSections = append(tmpAllSections[1:], curSec)

		stage2 := &sectionStage2{
			Name:     "",
			Content:  "",
			Size:     0,
			Children: []*sectionStage2{},
		}

		for _, s := range tmpAllSections {
			if len(s.Content) != 0 {
				s.Content = s.Content[1:]
			}

			tmp := &sectionStage2{
				Name:     s.Name,
				Content:  s.Content,
				Size:     s.Size,
				Parent:   stage2,
				Children: []*sectionStage2{},
			}

			if s.Size > stage2.Size {
				// h1 should not have h3 under it.
				if stage2.Size+1 != s.Size {
					continue
				}
			} else {
				for s.Size < stage2.Size {
					stage2 = stage2.Parent
				}

				stage2 = stage2.Parent
			}

			stage2.Children = append(stage2.Children, tmp)
			stage2 = tmp
		}

		for stage2.Parent != nil {
			stage2 = stage2.Parent
		}

		sections = stage2.Children

		newSections := []*sectionStage2{}

		for _, sec := range sections {
			if sec.Content != "" {
				newSections = append(newSections, sec)
				continue
			}

			f := false

			for _, cmd := range sec.Children {
				if discord.Command(cmd.Name) != nil {
					f = true
					break
				}
			}

			if f {
				newSections = append(newSections, sec)
			}
		}

		sections = newSections

		if len(sections) == 0 {
			return
		}

		opts := []*discordgo.ApplicationCommandOptionChoice{}

		for _, sec := range sections {
			opts = append(opts, &discordgo.ApplicationCommandOptionChoice{
				Name:  sec.Name,
				Value: sec.Name,
			})
		}

		discord.RegisterCommand(&discordgo.ApplicationCommand{
			Type:              discordgo.ChatApplicationCommand,
			Name:              "help",
			DefaultMemberPermissions: discord.Perms(),
			Description:       "Show the help menu",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "section",
					Required:    false,
					Description: "The section you want to quick-jump to",
					Choices:     opts,
				},
			},
		}, func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.ApplicationCommandInteractionData, data map[string]*discordgo.ApplicationCommandInteractionDataOption) {
			secName := sections[0].Name
			if data["section"] != nil {
				secName = data["section"].StringValue()
			}

			helpResp(s, i.Interaction, secName, false)
		})
		discord.RegisterComponent("help", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
			spl := strings.Split(d.CustomID, "_")[1:]

			action := spl[0]
			curI, _ := strconv.Atoi(spl[1])

			if action == "next" {
				curI++
			} else {
				curI--
			}

			helpResp(s, i.Interaction, sections[curI].Name, true)
		})
	}, &initializer.ModuleInfo{
		ShouldLoad: func(c *initializer.InitContext) bool {
			b, err := os.ReadFile("resources/help.md")
			return err == nil && len(b) != 0
		},
		PreHooks: []initutils.Module{initializer.MOD_DISCORD},
	})

	initializer.Register(initializer.MOD_HELP, func(c *initializer.InitContext) {
		for _, section := range sections {
			secDocs := []string{}

			str := section.Content + "\n"

			for _, cmd := range section.Children {
				tmp := cmd.CommandString(cmd.Name, "")
				if tmp == "" {
					continue
				}
				tmp += "\n"

				if len(str)+len(tmp) > 4000 {
					secDocs = append(secDocs, str)
					str = tmp
				}
			}

			secDocs = append(secDocs, str)

			newDocs := []string{}

			for _, v := range secDocs {
				tmp := strings.TrimSpace(v)
				if tmp == "" {
					continue
				}

				newDocs = append(newDocs, tmp)
			}

			if len(newDocs) == 0 {
				continue
			}

			sectionMap[section.Name] = newDocs
		}
	}, &initializer.ModuleInfo{
		ShouldLoad: func(c *initializer.InitContext) bool {
			return len(sections) != 0
		},
	}, initializer.MOD_DISCORD)
}
