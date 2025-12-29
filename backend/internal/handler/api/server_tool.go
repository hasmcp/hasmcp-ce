package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathServerTools      = _routePathServers + "/:id/tools"
	_routePathCreateServerTool = _routePathServerTools
	_routePathListServerTools  = _routePathServerTools
	_routePathDeleteServerTool = _routePathServerTools + "/:toolID"
)

func (h *handler) registerServerToolRoutes() error {
	h.router.Post(_routePathCreateServerTool, h.createServerTool())
	h.router.Get(_routePathListServerTools, h.listServerTools())
	h.router.Delete(_routePathDeleteServerTool, h.deleteServerTool())

	return nil
}

func (h *handler) createServerTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToCreateServerToolRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateServerTool(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateServerToolResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listServerTools() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToListServerToolsRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListServerTools(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListServerToolsResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) deleteServerTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToDeleteServerToolRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteServerTool(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
