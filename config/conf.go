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
		Port:     "3000",
		Disabled: map[string]bool{},
	}
	XP = xp{
		MsgMin: 16,
		MsgMax: 25,

		VCMin: 6,
		VCMax: 10,

		VCVideoOrStream:  1.5,
		VCVideoAndStream: 1.75,
		VCMute:           0.15,
		VCAlone:          0.15,
	}
	Games = games{
		Disable: map[string]bool{},
	}
)

/*
The conf:"" struct tag for dummies (and regular people, this is custom made for this project)

As is standard, each value is comma separated. Not all keys are required. With the exception of the 'consequence', everything else MUST be lowercase.

1st value must always be present. It represents the key used in the .conf file (KEY_HERE="VALUE HERE")
2nd value is not required, but will always be the 2nd value. It represents the module this key relates to (see initializer's DMOD modules). It can be left empty for a 'general' module
3d  value is not required & can be replaced. This can represent either the 'required' status for the key (set this value to be 'required' to make it a required key), or the consequence (more on this later). If there are 4 values, it will only be used as a requirement check, if there are 3 values and this is not set to 'required', then this will be used as the consequence
4th value is not required & cannot be replaced. If present, this will always represent the 'consequence' for leaving this value empty. The format is 'Due to X not being set, {CONSEQUENCE}'. It is recommended to keep this as consistent as possible with other values so that it can be grouped!
*/

type general struct {
	DB       string                   `conf:"db_uri,,required"`
	Port     string                   `conf:"port"`
	Domain   string                   `conf:"domain,twitch,twitch module will not be loaded"`
	Disabled caseInsensitiveInclusion `conf:"disable"`
}

type discord struct {
	GuildID   string   `conf:"guild,,required"`
	Token     string   `conf:"token,,required"`
	InviteURL string   `conf:"invite_url,twitch,the !discord command will not be available on twitch"`
	AutoRoles []string `conf:"auto_roles,auto_roles,there will be no roles added on user join"`
	OwnerID   string   `conf:"owner_id,,owner-specific features won't work :("`
}

type channels struct {
	XPAnnouncements string `conf:"xp_announcements,xp,xp level changes will not be announced"`
	Polls           string `conf:"polls,polls,the polls module will not be loaded"`
	RoleAssignment  string `conf:"role_assignment,toggle_roles,the role assignment module will not be loaded"`
	Finland         string `conf:"finland,finland,no finland :("`
	ModLog          string `conf:"mod_log,mod_log,warnings and message edits/deletes will not be logged"`
	Twitch          string `conf:"twitch,twitch,twitch streams will not be announced"`
}

type warnings struct {
	Punishments     *WarningLevels `conf:"punishments,,the only punishment will be the auto ban after severity"`
	AutoBanSeverity int            `conf:"auto_ban_severity"`
}

type donations struct {
	Donations  string         `conf:"channel_donations,donation,new donations will not be announced"`
	Funds      string         `conf:"channel_funds,donations,new funds will not be announced"`
	Info       string         `conf:"channel_info,donations,donation tier info will not posted"`
	Persistent *DonationRoles `conf:"roles_persistent,donations,there are no permanent roles for donors"`
	Monthly    *DonationRoles `conf:"roles_monthly,donations,there are no special roles for this month's donors"`
	Token      string         `conf:"token,donations,required"`
	Location   string         `conf:"location,donations"`
}

type twitch struct {
	ClientID     string `conf:"client_id,twitch,twitch module will not be loaded"`
	ClientSecret string `conf:"client_secret,twitch,twitch module will not be loaded"`
	AppName      string `conf:"app_name,twitch,twitch module will not be loaded"`
	ChannelName  string `conf:"channel_name,twitch,twitch module will not be loaded"`
	CustomSecret string `conf:"custom_secret,twitch,twitch module will not be loaded"`
	Role         string `conf:"ping_role,twitch,no one will be pinged about a new twitch stream"`
	CustomSO     map[string]string
	OwnID        string
}

type xp struct {
	MsgMin int `conf:"msg_min,xp"`
	MsgMax int `conf:"msg_max,xp"`

	VCMin int `conf:"vc_min,xp"`
	VCMax int `conf:"vc_max,xp"`

	VCVideoOrStream  float64 `conf:"vc_video,xp"`
	VCVideoAndStream float64 `conf:"vc_video_2,xp"`
	VCMute           float64 `conf:"vc_mute,xp"`
	VCAlone          float64 `conf:"vc_alone,xp"`
}

func (t twitch) ShouldLoad() bool {
	opts := []string{
		General.Domain,
		t.ClientID,
		t.ClientSecret,
		t.AppName,
		t.ChannelName,
		t.CustomSecret,
	}

	for _, o := range opts {
		if o == "" {
			return false
		}
	}

	return true
}

type debug struct {
	Mention string `conf:"mention"`
	WebHook string `conf:"webhook,,warnings & errors will not be sent to anything on discord"`
}

type games struct {
	Disable caseInsensitiveInclusion `conf:"disable,games"`
}
