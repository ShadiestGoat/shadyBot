package config

var (
	debugV = debug{}
	Twitch = twitch{
		CustomSO: map[string]string{},
	}
	Donations = donations{
		Persistent: new(DonationRoles),
		Monthly:    new(DonationRoles),
		Location:   "donate.shadygoat.eu",
	}
	Warnings = warnings{
		AutoBanSeverity: 5,
		Punishments:     new(WarningLevels),
	}
	Channels = channels{}
	Discord  = discord{}
	General  = general{
		Port: "3000",
	}
)

type general struct {
	DB     string `conf:"db_uri,required"`
	Port   string `conf:"port"`
	Domain string `conf:"domain,twitch module will not be loaded"`
}

type discord struct {
	GuildID   string   `conf:"guild,required"`
	Token     string   `conf:"token,required"`
	InviteURL string   `conf:"invite_url,the !discord command will not be available on twitch"`
	AutoRoles []string `conf:"auto_roles,there will be no roles added on user join"`
}

type channels struct {
	XPAnnouncements string `conf:"xp_announcements,xp level changes will not be announced"`
	Polls           string `conf:"polls,the polls module will not be loaded"`
	RoleAssignment  string `conf:"role_assignment,the role assignment module will not be loaded"`
	Finland         string `conf:"finland,no finland"`
	ModLog          string `conf:"mod_log,warnings and message edits/deletes will not be logged"`
	Twitch          string `conf:"twitch,twitch streams will not be announced"`
}

type warnings struct {
	Punishments     *WarningLevels `conf:"punishments,the only punishment will be the auto ban after severity"`
	AutoBanSeverity int            `conf:"auto_ban_severity"`
}

type donations struct {
	Donations  string         `conf:"channel_donations,new donations will not be announced"`
	Funds      string         `conf:"channel_funds,new funds will not be announced"`
	Info       string         `conf:"channel_info,donation tier info will not posted"`
	Persistent *DonationRoles `conf:"roles_persistent,there are no permanent roles for donors"`
	Monthly    *DonationRoles `conf:"roles_monthly,there are no special roles for this month's donors"`
	Token      string         `conf:"token,required"`
	Location   string         `conf:"location"`
}

type twitch struct {
	ClientID     string `conf:"client_id,twitch module will not be loaded"`
	ClientSecret string `conf:"client_secret,twitch module will not be loaded"`
	AppName      string `conf:"app_name,twitch module will not be loaded"`
	ChannelName  string `conf:"channel_name,twitch module will not be loaded"`
	CustomSecret string `conf:"custom_secret,twitch module will not be loaded"`
	Role         string `conf:"role,no one will be pinged about a new twitch stream"`
	CustomSO     map[string]string
}

type debug struct {
	Mention string `conf:"mention"`
	WebHook string `conf:"webhook,warnings & errors will not be sent to anything on discord"`
}
