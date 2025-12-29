package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/api/middleware/auth"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/api/middleware/cors"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/api/middleware/logger"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/api/middleware/ratelimit"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/server"
)

type (
	Params struct {
		Config config.Service
		Server server.Service
		Crud   crud.Controller
	}

	Handler interface {
	}

	handler struct {
		crud   crud.Controller
		router fiber.Router
	}
)

const (
	_routeBasePath = "/api/v1"

	headerContentType                     = "content-type"
	headerContentTypeValueApplicationJSON = "application/json"
)

var (
	_invalidRequestPayloadHTTPError = []byte(`{"error": {"message":"invalid request payload", "code":400}}`)
)

func New(p Params) (Handler, error) {
	auth, err := auth.New(auth.Params{
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}

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

	router := p.Server.Group(_routeBasePath).Use(logger).Use(cors).Use(ratelimit).Use(auth)
	h := &handler{
		router: router,
		crud:   p.Crud,
	}
	if err := h.registerVariableRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerProviderToolRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerProviderRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerServerToolRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerServerTokenRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerServerRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerResourceRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerPromptRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerServerPromptAssociationRoutes(); err != nil {
		return nil, err
	}

	if err := h.registerServerResourceAssociationRoutes(); err != nil {
		return nil, err
	}

	return h, nil
}
