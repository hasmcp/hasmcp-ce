package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathResources      = "/resources"
	_routePathCreateResource = _routePathResources
	_routePathListResources  = _routePathResources
	_routePathGetResource    = _routePathResources + "/:id"
	_routePathPatchResource  = _routePathResources + "/:id"
	_routePathDeleteResource = _routePathResources + "/:id"
)

func (h *handler) registerResourceRoutes() error {
	h.router.Post(_routePathCreateResource, h.createResource())
	h.router.Get(_routePathListResources, h.listResources())
	h.router.Get(_routePathGetResource, h.getResource())
	h.router.Patch(_routePathPatchResource, h.updateResource())
	h.router.Delete(_routePathDeleteResource, h.deleteResource())

	return nil
}

func (h *handler) createResource() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToCreateResourceRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateResource(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateResourceResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listResources() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToListResourcesRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListResources(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListResourcesResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) getResource() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToGetResourceRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.GetResource(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromGetResourceResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) updateResource() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToUpdateResourceRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.UpdateResource(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}

func (h *handler) deleteResource() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToDeleteResourceRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteResource(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
