# Shady Bot

This is the bot run on [my twitch](https://twitch.tv/shadiestgoat) and [my discord](https://discord.gg/Nq9vn6j3PS). This mainly focuses on discord however.

This project is organized in 'modules', which are sub packages (folders) in here. Here is the current list of modules:

| module | description | 
|:------:|-------------|
| config | The stuff that loads all your config values (read later sections) |
| auto roles | Add roles upon a member joining the discord |
| connect4 | A connect 4 game |
| donation | A module that connects [my donation website](https://donate.shadygoat.eu) to discord. [Github](https://github.com/ShadiestGoat/donations) |
| finland | Send some love to finland |
| games | ~~thinly disguised gambling~~ this is a group of games where users can bet their discord XP while playing several luck & skill based games (blackjack, coinflip, slots)
| help | a module that will automatically create a /help command for this bot |
| mod log | a module that will create a mod log - warnings, message updates & removals |
| polls | a tool for making polls on discord |
| pronouns | a [pronoundb](https://pronoundb.org) integration for discord & twitch |
| purge | a module for purging entire channels |
| role assign | create messages with components to toggle a user's roles, can be used to sign up to certain notifications, etc |
| twitch | the biggest module - this will create live twitch notifications, run a chat bot with periodic messages, and add channel point rewards <3 |
| warnings | a warning/punishment/moderation system for the discord |
| xp | add discord user xp, with leaderboards etc |

# Config

This project is configured through a few files, which will be detailed in the following sections

## Main Config

The main portion of this project is configured through a `config.conf` and a `secrets.conf`. These `.conf` files are merged together, and bare no difference in between, its just if you want to have certain values which are more sensitive in 1 area just in case, this is where you can hide them in.

The `.conf` files unfortunately don't have a fully compliant parser, however, it is very workable. The main difference between this `.conf` and a standard `.conf`/`.ini` are the following:

- Single quotes are not supported, only unquoted or double quoted values are accepted
- inline comments are not supported, comments are only available on a new character
- `#` comments are not supported, only use `;`

For documentation on the values, check out `config.template.conf`. As a note, if a values is pre-filled in there, then that means it is the default value.

If you omit certain values or sections, that is fine, as long as they are not required. Please note, this may cause some modules to not load, but you will be warned about that when starting the app.

## Help Module

The text for the `/help` command is documented in `resources/help.md`. The format is 'markdown-ish'. Comments made with \<\!\-\- Comment Content --> will be ignored. Headings are `# Content`. These **must** be at the start of a new line! Empty lines will be normalized, so don't be afraid to play around with these <3. Lines which solely consist of `-` (like `---` or `----------`) will be ignored.

The help file is split by heading size, each heading size representing a different thing.

1. h1 (`#`) is used as a 'section name'. These are the sections visible in `/help {section}>`.
2. h2 (`##`) is used as the root command name. This is case & space sensitive! For each H2 under a H1, this will make a new entry. These will be replaced with a command mention
3. h3+ (`###` and more) are used for sub commands. It's fine to just not include these, if you don't have anything to document here.

Content that is not a heading (ie. a `p`) will be used as content for this section/command

Headings that are out of place, such as a `###` directly under a `#` will be ignored.

## Twitch

### Period Texts

These are messages that are sent out periodically. Each message is located on a new line, in `resources/twitch/periodic.md`. Empty lines are ignored. 

### Special Shoutouts

Sometimes, I shoutout a user very often. For this, I have a custom shoutout for them. The file `resources/twitch/so.md` contains all the custom shoutouts in the format of `name - shoutout text`. With this, on twitch if your run `!so @name` (or `!so name`) it will run output `shoutout text`. There is a default/fallback text for non-custom shoutouts!

# Running this

1. Configure!!!
   1. [Youll need a discord token from here](https://discord.com/developers)
   2. [You may need twitch client/app info from here](https://dev.twitch.tv/console/apps)
   3. [You may need a donation token, for which you need to setup your own donation site](https://github.com/ShadiestGoat/donations)
2. Download & Compile this code with `go build` (don't worry - build go apps is really fast & easy, this should take you less than a minute!)
3. Run the binary output
4. Profit

## Roadmap/What is left

- [ ] A config option to disable modules
  - [ ] Refactor the way module choose to load/not to load
- [ ] Refactor the way twitch is set up - there are 2 'modules' one pre-http and 1 post, and that should not be
