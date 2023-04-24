package utils

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

func Caps(str string) string {
	newStr := ""
	for _, s := range strings.Split(str, " ") {
		newStr += strings.ToUpper(s[:1]) + s[1:] + " "
	}
	return newStr[:len(newStr)-1]
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Random number generator, min & max are both included.
func RandInt(min, max int) int {
	v := rand.Intn(max+1-min) + min
	return v
}

// Do a 'chance' check, where the chance is the chance needed to pass (0-1)
func Chance(chance float64) bool {
	if chance <= 0 {
		return false
	}

	return rand.Float64() <= chance
}

func ParseColor(inp string) (int, error) {
	if len(inp) != 0 && inp[0] == '#' {
		inp = "0x" + inp[1:]
	}

	i, err := strconv.ParseInt(inp, 0, 0)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

var Alphabet = []string{
	"a",
	"b",
	"c",
	"d",
	"e",
	"f",
	"g",
	"h",
	"i",
	"j",
	"k",
	"l",
	"m",
	"n",
	"o",
	"p",
	"q",
	"r",
	"s",
	"t",
	"u",
	"v",
	"w",
	"x",
	"y",
	"z",
}

// TODO: move RankResp into XP
type RankResp struct {
	Img   io.Reader
	Embed *discordgo.MessageEmbed
}

func (r RankResp) Resp() *discordgo.WebhookEdit {
	if r.Img != nil {
		return &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        "rank.png",
					ContentType: "image/png",
					Reader:      r.Img,
				},
			},
		}
	} else if r.Embed != nil {
		return &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				r.Embed,
			},
		}
	}
	return nil
}

// Replaces a string template, where the variable is {{%var_name}}
func ParseTemplateString(inp string, vars map[string]string) string {
	items := []string{}
	for varName, val := range vars {
		items = append(items, "{{%"+varName+"}}")
		items = append(items, val)
	}
	r := strings.NewReplacer(items...)

	last := ""
	for last != inp {
		last = inp
		inp = r.Replace(inp)
	}

	return inp
}

// Creates a text progress bar. Goal is the target number, progress is the number out of goal.
//
// 70% progress bar of 200 characters is TextProgressBar(70, 100, "", "", 200), or TextProgressBar(0.7, 1, "", "", 200) or TextProgressBar(14, 20, "", "", 200)
func TextProgressBar(goal float64, progress float64, leftStr, rightStr string, size int) string {
	str := ""

	p := (progress / goal)
	if p > 1 {
		p = 1
	}
	if p < 0 {
		p = 0
	}

	epic := float64(size) * p

	for i := 0.0; i < epic; i++ {
		str += "█"
	}

	for utf8.RuneCountInString(str) != size {
		str += "─"
	}

	return fmt.Sprintf("`%v ├%v┤ %v`", leftStr, str, rightStr)
}
