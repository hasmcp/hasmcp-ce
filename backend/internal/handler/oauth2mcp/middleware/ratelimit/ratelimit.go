package ratelimit

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	zlog "github.com/rs/zerolog/log"
)

type (
	Params struct {
		Config config.Service
	}

	ratelimitConfig struct {
		Enabled  bool          `yaml:"enabled"`
		MaxPerIP int           `yaml:"maxPerIP"`
		Window   time.Duration `yaml:"window"`
	}
)

const (
	_cfgKey = "oauth2ratelimit"

	_logPrefix = "[middleware/oauth2ratelimit] "
)

var (
	_ratelimited = []byte(`{"error": {"message":"ratelimited", "code":429}}`)
)

func New(p Params) (fiber.Handler, error) {
	var cfg ratelimitConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Enabled {
		zlog.Info().Msg(_logPrefix + "enabled")
		return limiter.New(limiter.Config{
			Max:        cfg.MaxPerIP,
			Expiration: cfg.Window,
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(429).Send(_ratelimited)
			},
		}), nil
	}

	handler := func(c *fiber.Ctx) error {
		return c.Next()
	}

	return handler, nil
}
