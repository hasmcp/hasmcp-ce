package jwt

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

type (
	Params struct {
		Config  config.Service
		JWTAuth jwt.AuthenticatorController
	}

	jwtConfig struct {
	}

	ctxKey uint8
)

const (
	_cfgKey = "mcpjwtauth"

	_ctxKeyAuthResult ctxKey = iota

	_headerKeyAuthorization     = "x-hasmcp-key"
	_headerValuePrefixBearer    = "Bearer "
	_headerValuePrefixBearerLen = len(_headerValuePrefixBearer)
)

var (
	_unauthorized = []byte(`{"error": {"message":"authentication failed", "code":401}}`)
)

func New(p Params) (fiber.Handler, error) {
	var cfg jwtConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	auth := p.JWTAuth
	handler := func(c *fiber.Ctx) error {
		tokenVal := c.Get(_headerKeyAuthorization)
		if tokenVal == "" {
			tokenVal = c.Query("token")
		} else if len(tokenVal) < _headerValuePrefixBearerLen ||
			tokenVal[0:_headerValuePrefixBearerLen] != _headerValuePrefixBearer {
			return c.Status(401).Send(_unauthorized)
		}

		if len(tokenVal) < 1 {
			return c.Status(401).Send(_unauthorized)
		}

		auth, err := auth.Authenticate(context.Background(), jwt.AuthParams{
			AccessToken: []byte(tokenVal[_headerValuePrefixBearerLen:]),
		})
		if err != nil {
			zlog.Error().Err(err).Msg("auth failed")
			return c.Status(401).Send(_unauthorized)
		}

		path := c.Path()
		if len(path) < 16 {
			zlog.Error().Err(errors.New("invalid path")).Str("path", path).Msg("auth failed")
			return c.Status(401).Send(_unauthorized)
		}
		id := []byte(path[5:16])

		serverID := monoflake.IDFromBase62(string(id)).Int64()

		if auth.ServerID != serverID {
			zlog.Info().Str("actual", monoflake.ID(auth.ServerID).String()).
				Str("param", c.Params("id")).Msg("server id mismatch")
			return c.Status(401).Send(_unauthorized)
		}

		c.Locals(_ctxKeyAuthResult, *auth)

		return c.Next()
	}

	return handler, nil
}

func AuthResult(c *fiber.Ctx) jwt.AuthResult {
	res, ok := c.Locals(_ctxKeyAuthResult).(jwt.AuthResult)
	if !ok {
		return jwt.AuthResult{}
	}
	return res
}
