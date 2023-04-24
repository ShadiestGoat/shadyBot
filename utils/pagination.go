package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	ENTRIES_PER_PAGE = 8
	SUPER_NUM        = 5
	SUPER_NUM_SMALL  = 2
)

// prefix_action_curPage_extraInfo

func PageBounds(page int) (int, int) {
	return (page - 1) * ENTRIES_PER_PAGE, page * ENTRIES_PER_PAGE
}

func ParsePagination(customID string, superNum int) (extraInfo []string, newPage int) {
	cInfo := strings.Split(customID, "_")[1:]

	oldPage, _ := strconv.Atoi(cInfo[1])

	n := 0

	switch cInfo[0] {
	case "superprev":
		n = -superNum
	case "prev":
		n = -1
	case "next":
		n = 1
	case "supernext":
		n = superNum
	case "refresh":
		n = 0
	}

	newPage = oldPage + n

	extraInfo = cInfo[2:]

	return
}

func PaginationButtonFactory(curPage int, maxL int, prefix string, fullDisable bool, superNum int, extraInfo string) []discordgo.MessageComponent {
	idSuffix := fmt.Sprint(curPage)

	if extraInfo != "" {
		idSuffix += "_" + extraInfo
	}

	// if true, these buttons should be disabled
	superPrev, prev, next, superNext := fullDisable, fullDisable, fullDisable, fullDisable

	if !fullDisable {
		start, _ := PageBounds(curPage - superNum)
		if start <= -ENTRIES_PER_PAGE {
			superPrev = true
		}

		start, _ = PageBounds(curPage - 1)
		if start <= -ENTRIES_PER_PAGE {
			prev = true
		}

		start, _ = PageBounds(curPage + 1)
		if start >= maxL {
			next = true
		}

		start, _ = PageBounds(curPage + superNum)
		if start >= maxL {
			superNext = true
		}
	}

	return []discordgo.MessageComponent{
		&discordgo.Button{
			Disabled: superPrev,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: "⏮",
			},
			CustomID: prefix + "_superprev_" + idSuffix,
		},
		&discordgo.Button{
			Disabled: prev,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: "⬅️",
			},
			CustomID: prefix + "_prev_" + idSuffix,
		},
		&discordgo.Button{
			Disabled: false,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: "🔄",
			},
			CustomID: prefix + "_refresh_" + idSuffix,
		},
		&discordgo.Button{
			Disabled: next,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: "➡️",
			},
			CustomID: prefix + "_next_" + idSuffix,
		},
		&discordgo.Button{
			Disabled: superNext,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: "⏭️",
			},
			CustomID: prefix + "_supernext_" + idSuffix,
		},
	}
}
