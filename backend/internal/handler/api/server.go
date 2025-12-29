package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"

	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathServers      = "/servers"
	_routePathCreateServer = _routePathServers
	_routePathListServers  = _routePathServers
	_routePathGetServer    = _routePathServers + "/:id"
	_routePathPatchServer  = _routePathServers + "/:id"
	_routePathDeleteServer = _routePathServers + "/:id"
)

func (h *handler) registerServerRoutes() error {
	h.router.Post(_routePathCreateServer, h.createServer())
	h.router.Get(_routePathListServers, h.listServers())
	h.router.Get(_routePathGetServer, h.getServer())
	h.router.Patch(_routePathPatchServer, h.updateServer())
	h.router.Delete(_routePathDeleteServer, h.deleteServer())

	return nil
}

func (h *handler) createServer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToCreateServerRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateServer(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateServerResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listServers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToListServersRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListServers(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListServersResponseEntityToHTTPResponse(rs)
		return c.Send(payload)
	}
}

func (h *handler) getServer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToGetServerRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.GetServer(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromGetServerResponseEntityToHTTPResponse(rs)
		return c.Send(payload)
	}
}

func (h *handler) updateServer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToUpdateServerRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.UpdateServer(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromUpdateServerResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) deleteServer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToDeleteServerRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteServer(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
