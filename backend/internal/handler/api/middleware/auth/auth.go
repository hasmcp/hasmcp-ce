package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	zlog "github.com/rs/zerolog/log"
)

type (
	Params struct {
		Config config.Service
	}

	authConfig struct {
		Enabled        bool   `yaml:"enabled"`
		ApiAccessToken string `yaml:"apiAccessToken"`
	}

	ctxKey uint8
)

const (
	_cfgKey = "apiauth"

	_ctxKeyAuthResult ctxKey = iota

	_logPrefix = "[middleware/apiauth] "
)

var (
	_unauthorized = []byte(`{"error": {"message":"authentication failed", "code":401}}`)
)

func New(p Params) (fiber.Handler, error) {
	var cfg authConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Enabled {
		zlog.Info().Msg(_logPrefix + "enabled")
	}

	apiAccessToken := "Bearer " + cfg.ApiAccessToken
	enabled := cfg.Enabled

	handler := func(c *fiber.Ctx) error {
		if !enabled {
			return c.Next()
		}
		if c.Method() == "OPTIONS" {
			return c.Next()
		}
		if c.Get("Authorization") == apiAccessToken {
			return c.Next()
		}
		return c.Status(401).Send(_unauthorized)
	}

	return handler, nil
}
