package main

import (
	"os"
	"os/signal"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/shadyBot/initializer"

	_ "github.com/shadiestgoat/shadyBot/autoroles"
	_ "github.com/shadiestgoat/shadyBot/characters"
	_ "github.com/shadiestgoat/shadyBot/config"
	_ "github.com/shadiestgoat/shadyBot/connect4"
	_ "github.com/shadiestgoat/shadyBot/db"
	_ "github.com/shadiestgoat/shadyBot/discord"
	_ "github.com/shadiestgoat/shadyBot/discutils"
	_ "github.com/shadiestgoat/shadyBot/donation"
	_ "github.com/shadiestgoat/shadyBot/finland"
	_ "github.com/shadiestgoat/shadyBot/games"
	_ "github.com/shadiestgoat/shadyBot/help"
	_ "github.com/shadiestgoat/shadyBot/initializer"
	_ "github.com/shadiestgoat/shadyBot/misc"
	_ "github.com/shadiestgoat/shadyBot/modLog"
	_ "github.com/shadiestgoat/shadyBot/polls"
	_ "github.com/shadiestgoat/shadyBot/pronouns"
	_ "github.com/shadiestgoat/shadyBot/purge"
	_ "github.com/shadiestgoat/shadyBot/roleAssign"
	_ "github.com/shadiestgoat/shadyBot/router"
	_ "github.com/shadiestgoat/shadyBot/snownode"
	_ "github.com/shadiestgoat/shadyBot/twitch"
	_ "github.com/shadiestgoat/shadyBot/utils"
	_ "github.com/shadiestgoat/shadyBot/warnings"
	_ "github.com/shadiestgoat/shadyBot/xp"
	_ "github.com/shadiestgoat/shadyBot/xpCommands"
)

func main() {
	initializer.Init()
	defer initializer.Close()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)

	log.Success("Everything should be loaded up!")

	log.PrintDebug("You can now use Ctrl+C to stop this application!")

	<-c
}
