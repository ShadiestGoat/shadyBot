package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/shadiestgoat/shadyBot/config"
	"github.com/shadiestgoat/shadyBot/initializer"
)

var r = chi.NewRouter()
var s = &http.Server{}

func init() {
	r.Use(middleware.StripSlashes, middleware.CleanPath, middleware.URLFormat, middleware.NoCache, cors.AllowAll().Handler)

	initializer.Register(initializer.MOD_HTTP, func(c *initializer.InitContext) {
		s.Addr = ":" + config.General.Port
		s.Handler = r

		go s.ListenAndServe()
	}, nil)

	initializer.RegisterCloser(initializer.MOD_HTTP, func() {
		c, cancel := context.WithTimeout(context.Background(), 8*time.Second)

		s.Shutdown(c)

		cancel()
	})
}

// Register a route on root router
func Register(route string, h http.Handler) {
	r.Mount(route, h)
}

func RegisterCloser(h func()) {
	s.RegisterOnShutdown(h)
}
