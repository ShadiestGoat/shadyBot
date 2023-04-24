package donation

import (
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
)

func donationRolesRaw(donated float64, curRoles []string, roles []*config.DonationRole) []string {
	newRoles := []string{}

	parsedRoles := map[string]bool{}

	for _, r := range roles {
		parsedRoles[r.RoleID] = true
	}

	for _, roleID := range curRoles {
		if !parsedRoles[roleID] {
			newRoles = append(newRoles, roleID)
		}
	}

	for _, r := range roles {
		if r.Min < donated && (r.Max == -1 || donated <= r.Max) {
			newRoles = append(newRoles, r.RoleID)
		}
	}

	return newRoles
}

func setDonationRoles(s *discordgo.Session, c *donations.Client, userID string, curRoles []string) []string {
	newRoles := make([]string, len(curRoles))
	copy(newRoles, curRoles)

	resp, err := c.DonorByDiscord(userID, false)
	if log.ErrorIfErr(err, "fetching donor discord '%s' user", userID) || resp == nil {
		if err.(*donations.HTTPError).Status == 404 {
			return curRoles
		}
		time.Sleep(2 * time.Second)
		return setDonationRoles(s, c, userID, curRoles)
	}

	newRoles = donationRolesRaw(resp.Total.Total, newRoles, *config.Donations.Persistent)
	newRoles = donationRolesRaw(resp.Total.Month, newRoles, *config.Donations.Monthly)
	editRoles := true

	if len(newRoles) == len(curRoles) {
		sort.StringSlice(curRoles).Sort()
		sort.StringSlice(newRoles).Sort()

		found := false
		for i, v := range curRoles {
			if newRoles[i] != v {
				found = true
				break
			}
		}

		editRoles = found
	}

	if editRoles {
		_, err := s.GuildMemberEditComplex(config.Discord.GuildID, userID, &discordgo.GuildMemberParams{
			Roles: &newRoles,
		})

		if log.ErrorIfErr(err, "setting member roles after don") {
			return setDonationRoles(s, c, userID, curRoles)
		}
	}

	if resp.Total.Month > 0 {
		donationQueue.Add(userID)
	}

	return newRoles
}
