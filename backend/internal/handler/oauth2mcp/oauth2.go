package oauth2mcp

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	oauth2 "github.com/hasmcp/hasmcp-ce/backend/internal/controller/oauth2mcp"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/oauth2mcp/middleware/cors"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/oauth2mcp/middleware/logger"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/oauth2mcp/middleware/ratelimit"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/oauth2mcp"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/server"
	"github.com/valyala/fasthttp"
)

type (
	Params struct {
		Config config.Service
		Server server.Service
		Oauth2 oauth2.Controller
	}

	Handler interface {
	}

	handler struct {
		oauth2 oauth2.Controller
		router fiber.Router
	}
)

const (
	_routeBasePath         = "/oauth2"
	_routePathGetAuthorize = "/authorize"
	_routePathGetCallback  = "/callback"

	headerContentType                     = "content-type"
	headerContentTypeValueApplicationJSON = "application/json"
)

var (
	_invalidRequestPayloadHTTPError = []byte(`{"error": {"message":"invalid request payload", "code":400}}`)
)

func New(p Params) (Handler, error) {
	logger, err := logger.New(logger.Params{
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}

	ratelimit, err := ratelimit.New(ratelimit.Params{
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}

	cors, err := cors.New(cors.Params{
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}

	router := p.Server.Group(_routeBasePath).Use(logger).Use(ratelimit).Use(cors)
	h := &handler{
		router: router,
		oauth2: p.Oauth2,
	}
	if err := h.registerOauth2Routes(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *handler) registerOauth2Routes() error {
	h.router.Get(_routePathGetAuthorize, h.authorize())
	h.router.Get(_routePathGetCallback, h.callback())

	return nil
}

func (h *handler) authorize() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rq := mapper.FromHTTPRequestToOauth2AuthorizeRequestEntity(c)
		if rq == nil {
			c.Set(headerContentType, headerContentTypeValueApplicationJSON)
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.oauth2.Authorize(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Set(headerContentType, headerContentTypeValueApplicationJSON)
			c.Status(status)
			return c.Send(e)
		}

		return c.Redirect(rs.AuthCodeURL, fasthttp.StatusTemporaryRedirect)
	}
}

func (h *handler) callback() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rq := mapper.FromHTTPRequestToOauth2CallbackRequestEntity(c)
		if rq == nil {
			c.Set(headerContentType, headerContentTypeValueApplicationJSON)
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.oauth2.Callback(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Set(headerContentType, headerContentTypeValueApplicationJSON)
			c.Status(status)
			return c.Send(e)
		}

		return c.Redirect(rs.InternalRedirectURL, fasthttp.StatusTemporaryRedirect)
	}
}
