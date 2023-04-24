package donation

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discutils"
	"github.com/shadiestgoat/shadyBot/utils"
)

type DonationQueueStore struct {
	*sync.Mutex
	Queue []string
}

var donationQueue = &DonationQueueStore{
	Mutex: &sync.Mutex{},
	Queue: []string{},
}

func (s *DonationQueueStore) Add(userID string) {
	s.Lock()
	defer s.Unlock()

	if utils.BinarySearch(s.Queue, userID) != -1 {
		return
	}

	s.Queue = append(s.Queue, userID)
}

func (store *DonationQueueStore) Loop(s *discordgo.Session, c *donations.Client) {
	for {
		store.Lock()

		for _, userID := range store.Queue {
			m := discutils.GetMember(s, config.Discord.GuildID, userID)
			if m == nil {
				continue
			}
			go setDonationRoles(s, c, m.User.ID, m.Roles)
		}

		store.Unlock()
		time.Sleep(12 * time.Hour)
	}
}
