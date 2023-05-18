package initializer

import (
	"fmt"
	"sync"

	"github.com/shadiestgoat/initutils"
	"github.com/shadiestgoat/log"

	// ... these deps are used for context, so they will be different for your application
	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
)

const (
	MOD_AUTO_ROLES        initutils.Module = "AUTO_ROLES"
	MOD_ROLE_ASSIGN       initutils.Module = "ROLE_ASSIGN"
	MOD_CONFIG            initutils.Module = "CONFIG"
	MOD_DB                initutils.Module = "DB"
	MOD_DISCORD           initutils.Module = "DISCORD"
	MOD_DONATION          initutils.Module = "DONATION"
	MOD_FINLAND           initutils.Module = "FINLAND"
	MOD_GAMES             initutils.Module = "GAMES"
	MOD_GAMBLER           initutils.Module = "GAMBLER"
	MOD_LOG               initutils.Module = "LOG"
	MOD_POLLS             initutils.Module = "POLLS"
	MOD_XP                initutils.Module = "XP"
	MOD_MOD_LOG           initutils.Module = "MOD_LOG"
	MOD_HELP_LOADER       initutils.Module = "HELP_LOADER"
	MOD_HELP              initutils.Module = "HELP"
	MOD_TWITCH            initutils.Module = "TWITCH"
	MOD_TWITCH_LIVE       initutils.Module = "TWITCH_LIVE_NOTIFICATION"
	MOD_TWITCH_CUSTOM_CMD initutils.Module = "TWITCH_CUSTOM_CMD"
	MOD_TWITCH_PERIODIC   initutils.Module = "TWITCH_PERIODIC"
	MOD_HTTP              initutils.Module = "HTTP"
	MOD_XP_CMD            initutils.Module = "XP_CMD"
)

var aliases = map[string][]initutils.Module{
	"xp":                    {MOD_XP, MOD_XP_CMD},
	"auto_roles":            {MOD_AUTO_ROLES},
	"toggle_roles":          {MOD_ROLE_ASSIGN},
	"donation":              {MOD_DONATION},
	"finland":               {MOD_FINLAND},
	"games":                 {MOD_GAMES},
	"polls":                 {MOD_POLLS},
	"mod_log":               {MOD_MOD_LOG},
	"help":                  {MOD_HELP_LOADER, MOD_HELP},
	"twitch":                {MOD_TWITCH, MOD_TWITCH_LIVE, MOD_TWITCH_CUSTOM_CMD},
	"twitch_live":           {MOD_TWITCH_LIVE},
	"twitch_custom_cmd":     {MOD_TWITCH_CUSTOM_CMD},
	"twitch_periodic_texts": {MOD_TWITCH_PERIODIC},
}

type InitContext struct {
	Discord  *discordgo.Session
	Donation *donations.Client
}

var ctx = &InitContext{}

var priorityInit = initutils.NewInitializer(ctx)
var normalInit = initutils.NewInitializer(ctx)

func RegisterPriority(m initutils.Module, h func(c *InitContext), preHooks []initutils.Module, dependencies ...initutils.Module) {
	priorityInit.Register(m, h, preHooks, dependencies...)
}

type ModuleInfo struct {
	ConfigOpts []*string
	// Custom function to determine if the module should be loaded. This should return **false** if the module should not be loaded.
	ShouldLoad func(c *InitContext) bool

	PreHooks []initutils.Module
}

var doNotClose = map[initutils.Module]bool{}

func Register(m initutils.Module, h func(c *InitContext), config *ModuleInfo, dependencies ...initutils.Module) {
	if config == nil {
		config = &ModuleInfo{
			ConfigOpts: []*string{},
			PreHooks:   []initutils.Module{},
		}
	}

	normalInit.Register(m, func(c *InitContext) {
		shouldLoad := true
		notLoadReason := ""

		if config.ShouldLoad != nil {
			shouldLoad = config.ShouldLoad(c)
			notLoadReason = "ShouldLoad() outputting false"
		} else {
			for i, conf := range config.ConfigOpts {
				if conf == nil || *conf == "" {
					shouldLoad = false
					notLoadReason = "missing config [" + fmt.Sprint(i) + "]"
					break
				}
			}
		}

		doNotClose[m] = !shouldLoad

		if shouldLoad {
			h(c)
			log.Success("[✅] Loaded the '%s' module!", string(m))
		} else {
			log.Warn("[❌] Didn't load the the '%s' module, due to %s!", string(m), notLoadReason)
		}
	}, config.PreHooks, dependencies...)
}

func Init() {
	err := priorityInit.Init()
	if err != nil {
		panic(err)
	}

	plan, err := normalInit.Plan()
	log.FatalIfErr(err, "creating a plan for normal modules")

	str := "Init plan:\n"

	for _, m := range plan {
		str += "- " + string(m) + "\n"
	}

	str = str[:len(str)-1]

	log.Debug(str)

	// same errors as normalInit.Plan()
	normalInit.Init()
}

var closers = map[initutils.Module]func(){}
var closeLock = &sync.Mutex{}

func RegisterCloser(mod initutils.Module, h func()) {
	closeLock.Lock()
	defer closeLock.Unlock()
	closers[mod] = h
}

func Close() {
	closeLock.Lock()
	for m, h := range closers {
		if doNotClose[m] {
			continue
		}

		h()
	}
}
