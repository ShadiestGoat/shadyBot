package roleassign

import (
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func init() {
	initializer.Register(initializer.MOD_ROLE_ASSIGN, func(c *initializer.InitContext) {
		register()
	}, &initializer.ModuleInfo{
		ConfigOpts: []*string{&config.Channels.Polls},
		PreHooks:   []initutils.Module{
			initializer.MOD_HELP_LOADER,
			initializer.MOD_DISCORD,
		},
	})
}
