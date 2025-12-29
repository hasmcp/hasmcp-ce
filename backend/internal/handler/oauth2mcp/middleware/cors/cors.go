package cors

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	zlog "github.com/rs/zerolog/log"
)

type (
	Params struct {
		Config config.Service
	}

	corsConfig struct {
		Enabled   bool     `yaml:"enabled"`
		Hostnames []string `yaml:"hostnames"`
	}
)

const (
	_cfgKey = "oauth2cors"

	_logPrefix = "[middleware/oauth2cors] "
)

func New(p Params) (fiber.Handler, error) {
	var cfg corsConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	allowedHostnames := map[string]struct{}{}
	for _, h := range cfg.Hostnames {
		allowedHostnames[h] = struct{}{}
	}

	if cfg.Enabled {
		zlog.Info().Msg(_logPrefix + "enabled")
		return cors.New(cors.Config{
			// Use a function to dynamically check the origin
			AllowOriginsFunc: func(origin string) bool {
				// An empty origin means it's not a cross-origin request
				if origin == "" {
					return true
				}

				// Parse the origin URL
				u, err := url.Parse(origin)
				if err != nil {
					return false
				}

				hostname := u.Hostname()
				if _, ok := allowedHostnames[hostname]; ok {
					return ok
				}
				if _, ok := allowedHostnames["*"]; ok {
					return ok
				}

				return false
			},
			// You can also specify other options, like allowed methods
			AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
			AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		}), nil
	}

	handler := func(c *fiber.Ctx) error {
		return c.Next()
	}

	return handler, nil
}
