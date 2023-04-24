package polls

import (
	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/initializer"
)

func init() {
	initializer.Register(initializer.MOD_POLLS, func(c *initializer.InitContext) {
		cmd()
		modal()
	}, &initializer.ModuleInfo{
		PreHooks: []initutils.Module{
			initializer.MOD_HELP_LOADER,
		},
		ConfigOpts: []*string{
			&config.Channels.Polls,
		},
	})
}
