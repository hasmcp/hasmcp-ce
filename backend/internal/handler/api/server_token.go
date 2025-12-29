package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathServerTokens      = _routePathServers + "/:id/tokens"
	_routePathCreateServerToken = _routePathServerTokens
)

func (h *handler) registerServerTokenRoutes() error {
	h.router.Post(_routePathCreateServerToken, h.createServerToken())

	return nil
}

func (h *handler) createServerToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToCreateServerTokenRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateServerToken(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateServerTokenResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}
