package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathProviderTools      = _routePathProviders + "/:id/tools"
	_routePathCreateProviderTool = _routePathProviderTools
	_routePathListProviderTools  = _routePathProviderTools
	_routePathGetProviderTool    = _routePathProviderTools + "/:toolID"
	_routePathPatchProviderTool  = _routePathProviderTools + "/:toolID"
	_routePathDeleteProviderTool = _routePathProviderTools + "/:toolID"
)

func (h *handler) registerProviderToolRoutes() error {
	h.router.Post(_routePathCreateProviderTool, h.createProviderTool())
	h.router.Get(_routePathListProviderTools, h.listProviderTools())
	h.router.Get(_routePathGetProviderTool, h.getProviderTool())
	h.router.Patch(_routePathPatchProviderTool, h.updateProviderTool())
	h.router.Delete(_routePathDeleteProviderTool, h.deleteProviderTool())
	return nil
}

func (h *handler) createProviderTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToCreateProviderToolRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateProviderTool(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateProviderToolResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listProviderTools() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToListProviderToolsRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}
		rs, err := h.crud.ListProviderTools(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListProviderToolsResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) getProviderTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToGetProviderToolRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.GetProviderTool(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromGetProviderToolResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) updateProviderTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToUpdateProviderToolRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.UpdateProviderTool(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromUpdateProviderToolResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) deleteProviderTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToDeleteProviderToolRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteProviderTool(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
