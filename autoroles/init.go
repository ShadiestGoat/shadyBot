package autoroles

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/discord"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func init() {
	initializer.Register(initializer.MOD_AUTO_ROLES, func(c *initializer.InitContext) {
		discord.MemberJoin.Add(func(s *discordgo.Session, v *discordgo.GuildMemberAdd) bool {
			// there might be earlier modules that edit this
			v.Member.Roles = append(v.Member.Roles, config.Discord.AutoRoles...)

			m, err := s.GuildMemberEdit(config.Discord.GuildID, v.User.ID, &discordgo.GuildMemberParams{
				Roles: &v.Member.Roles,
			})

			if !log.ErrorIfErr(err, "auto roles") {
				// just in case <3
				*v.Member = *m
			}

			return false
		})
	}, &initializer.ModuleInfo{
		ShouldLoad: func(_ *initializer.InitContext) bool {
			return len(config.Discord.AutoRoles) != 0
		},
		PreHooks: []initutils.Module{initializer.MOD_HELP_LOADER},
	})
}
