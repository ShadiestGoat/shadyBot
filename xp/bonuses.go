package xp

import (
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

func bonus(mem *discordgo.Member, curLVL int) int {
	if mem == nil {
		return 0
	}

	bonusChance := 1
	if discutils.IsMemberPremium(mem) {
		h := time.Since(*mem.PremiumSince).Hours()
		months := int(math.Ceil(h / (24 * 30)))
		bonusChance += months * 4
	}

	if bonusChance > utils.RandInt(0, 2500) {
		req := float64(LevelUpRequirement(curLVL))
		if req < 600 {
			req = 600.0
		}

		return utils.RandInt(int(math.Round(req*0.3)), int(math.Round(req*0.7)))
	}
	return 0
}
