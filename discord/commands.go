package discord

import "github.com/bwmarrin/discordgo"

var commands = map[string]*discordgo.ApplicationCommand{}

var t = true

// These are points to false/true values, do not change <3
var DefaultFalse = new(bool)
var DefaultTrue = &t

// Register a command and it's handler
func RegisterCommand(cmd *discordgo.ApplicationCommand, handler HandlerCommand) {
	if cmd.DefaultPermission == nil {
		cmd.DefaultPermission = DefaultTrue
	}
	if cmd.DefaultMemberPermissions == nil {
		cmd.DefaultMemberPermissions = Perms()
	}

	commands[cmd.Name] = cmd
	commandHandlers[cmd.Name] = handler
}

type Permission int64 

const (
	PERM_CREATE_INSTANT_INVITE Permission = 1 << iota
	PERM_KICK_MEMBERS
	PERM_BAN_MEMBERS
	PERM_ADMINISTRATOR
	PERM_MANAGE_CHANNELS
	PERM_MANAGE_GUILD
	PERM_ADD_REACTIONS
	PERM_VIEW_AUDIT_LOG
	PERM_PRIORITY_SPEAKER
	PERM_STREAM
	PERM_VIEW_CHANNEL
	PERM_SEND_MESSAGES
	PERM_SEND_TTS_MESSAGES
	PERM_MANAGE_MESSAGES
	PERM_EMBED_LINKS
	PERM_ATTACH_FILES
	PERM_READ_MESSAGE_HISTORY
	PERM_MENTION_EVERYONE
	PERM_USE_EXTERNAL_EMOJIS
	PERM_VIEW_GUILD_INSIGHTS
	PERM_CONNECT
	PERM_SPEAK
	PERM_MUTE_MEMBERS
	PERM_DEAFEN_MEMBERS
	PERM_MOVE_MEMBERS
	PERM_USE_VAD
	PERM_CHANGE_NICKNAME
	PERM_MANAGE_NICKNAMES
	PERM_MANAGE_ROLES
	PERM_MANAGE_WEBHOOKS
	PERM_MANAGE_GUILD_EXPRESSIONS
	PERM_USE_APPLICATION_COMMANDS
	PERM_REQUEST_TO_SPEAK
	PERM_MANAGE_EVENTS
	PERM_MANAGE_THREADS
	PERM_CREATE_PUBLIC_THREADS
	PERM_CREATE_PRIVATE_THREADS
	PERM_USE_EXTERNAL_STICKERS
	PERM_SEND_MESSAGES_IN_THREADS
	PERM_USE_EMBEDDED_ACTIVITIES
	PERM_MODERATE_MEMBERS
	PERM_VIEW_CREATOR_MONETIZATION_ANALYTICS
	PERM_USE_SOUNDBOARD
)

func Perms(perms ...Permission) *int64 {
	permsGiven := int64(PERM_SEND_MESSAGES)

	for _, p := range perms {
		permsGiven |= int64(p)
	}

	return &permsGiven
}

// Returns a command ID, or an empty string if the command is not found.
func CommandID(cmdName string) string {
	if cmd, ok := commands[cmdName]; ok {
		return cmd.ID
	}

	return ""
}

// Creates a command mention. subCommands should be a space separated sub command list (or empty).
// If the command does not exist, this will return an empty string.
func CommandMention(cmdName string, subCommands string) string {
	// </NAME SUBCOMMAND_GROUP SUBCOMMAND:ID>
	id := CommandID(cmdName)
	if id == "" {
		return ""
	}

	inside := cmdName
	if subCommands != "" {
		inside += " " + subCommands
	}

	return "</" + inside + ":" + id + ">"
}

func Command(name string) *discordgo.ApplicationCommand {
	return commands[name]
}
