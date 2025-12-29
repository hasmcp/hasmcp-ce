package static

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/server"
)

type (
	Params struct {
		Server server.Service
	}

	Handler interface {
	}

	handler struct {
	}
)

const (
	_routeBasePath     = "/"
	_routeWildcardPath = "/*"
)

func New(p Params) (Handler, error) {
	p.Server.Static(_routeBasePath, "./public")
	p.Server.Get(_routeWildcardPath, func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})

	return &handler{}, nil
}
