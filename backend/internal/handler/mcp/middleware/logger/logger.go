package logger

import (
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	zlog "github.com/rs/zerolog/log"
)

type (
	Params struct {
		Config config.Service
	}

	loggerConfig struct {
		Enabled bool `yaml:"enabled"`
	}
)

const (
	_cfgKey = "mcplogger"

	_logPrefix = "[middleware/mcplogger] "
)

func New(p Params) (fiber.Handler, error) {
	var cfg loggerConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Enabled {
		zlog.Info().Msg(_logPrefix + "enabled")
		return fiberzerolog.New(fiberzerolog.Config{
			Logger: &zlog.Logger,
		}), nil
	}

	handler := func(c *fiber.Ctx) error {
		return c.Next()
	}

	return handler, nil
}
