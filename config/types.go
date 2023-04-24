package config

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/utils"
)

type WarningLevels []*WarningLevel

func (wl *WarningLevels) Parse(inp string) error {
	spl := strings.Split(inp, "\\nl")
	curLvl := 1
	levels := []*WarningLevel{}

	for _, level := range spl {
		skipLvlR := ""
		for i, b := range level {
			if b >= '0' && b <= '9' {
				skipLvlR += string(b)
			} else {
				level = level[i:]
				break
			}
		}

		parsed, _ := strconv.Atoi(skipLvlR)
		curLvl = parsed

		lvl := &WarningLevel{
			MinWarnings: curLvl,
			Punishment:  WarningPunishment{},
		}
		splLvl := strings.Split(level, "\\np")

		for _, splLvl := range splLvl {
			switch splLvl[0] {
			case 'm':
				lvl.Punishment.Msg = splLvl[2:]
			case 't':
				t, err := strconv.Atoi(splLvl[2:])
				log.FatalIfErr(err, "parsing timeout time for %d warnings", curLvl)

				lvl.Punishment.Timeout = time.Duration(t) * time.Second
			case 'b':
				lvl.Punishment.Ban = true
			}
		}

		levels = append(levels, lvl)

		curLvl++
	}

	sort.Slice(levels, func(i, j int) bool {
		return levels[i].MinWarnings > levels[j].MinWarnings
	})

	*wl = WarningLevels(levels)

	return nil
}

type WarningLevel struct {
	MinWarnings int
	Punishment  WarningPunishment
}

func (p WarningPunishment) String() string {
	r := ""

	if p.Timeout != 0 {
		r = utils.FormatMinutes(int(math.Floor(p.Timeout.Minutes()))) + " Timeout"
	}

	if p.Ban {
		if r != "" {
			r += " & "
		}

		r += "Ban"
	}

	if r == "" {
		r = "A slap on the wrist"
	}

	return r
}

func (wl WarningLevel) String() string {
	return wl.Punishment.String()
}

type WarningPunishment struct {
	Ban     bool
	Timeout time.Duration
	Msg     string
}

type DonationRoles []*DonationRole

// Roles. min:max:xp_multiplier:role_id|min2:max2:xp_multiplier:role_id_2
// For donations. The min is non inclusive, so if min == 0, then non donors are not accepted into this category
// If max == -1, then there is no upper limit to it.
func (dr *DonationRoles) Parse(inp string) error {
	roles := []*DonationRole{}
	rolesOverall := strings.Split(inp, "|")

	for _, r := range rolesOverall {
		raw := strings.Split(r, ":")
		if len(raw) != 4 {
			log.Warn("Couldn't parse '%v': format should be 'min:max:xp_multiplier:role_id'", raw)
			continue
		}
		min, err := strconv.ParseFloat(raw[0], 64)
		if err != nil {
			log.Warn("Couldn't parse '%v': invalid float", err)
			continue
		}
		max, err := strconv.ParseFloat(raw[1], 64)
		if err != nil {
			log.Warn("Couldn't parse '%v': invalid float", err)
			continue
		}
		xpMult, err := strconv.ParseFloat(raw[2], 64)
		if err != nil {
			log.Warn("Couldn't parse '%v': invalid float", err)
			continue
		}
		roleID := raw[3]

		if (min < 0 || max < 0) && max != -1 {
			log.Warn("Role '%s' has min or max < 0, which is illegal!", roleID)
			continue
		}

		if max != -1 && min < max {
			log.Warn("Role '%s' min < max, which is very illegal (you're going to jail btw)!", roleID)
			continue
		}

		roles = append(roles, &DonationRole{
			Min:          min,
			Max:          max,
			XPMultiplier: xpMult,
			RoleID:       roleID,
		})
	}

	*dr = roles

	return nil
}

type DonationRole struct {
	Min          float64
	Max          float64
	XPMultiplier float64
	RoleID       string
}
