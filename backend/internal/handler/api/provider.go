package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathProviders      = "/providers"
	_routePathCreateProvider = _routePathProviders
	_routePathListProviders  = _routePathProviders
	_routePathGetProvider    = _routePathProviders + "/:id"
	_routePathPatchProvider  = _routePathProviders + "/:id"
	_routePathDeleteProvider = _routePathProviders + "/:id"
)

func (h *handler) registerProviderRoutes() error {
	h.router.Post(_routePathCreateProvider, h.createProvider())
	h.router.Get(_routePathListProviders, h.listProviders())
	h.router.Get(_routePathGetProvider, h.getProvider())
	h.router.Patch(_routePathPatchProvider, h.updateProvider())
	h.router.Delete(_routePathDeleteProvider, h.deleteProvider())

	return nil
}

func (h *handler) createProvider() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToCreateProviderRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateProvider(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateProviderResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listProviders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToListProvidersRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListProviders(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListProvidersResponseEntityToHTTPResponse(rs)
		return c.Send(payload)
	}
}

func (h *handler) getProvider() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToGetProviderRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.GetProvider(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromGetProviderResponseEntityToHTTPResponse(rs)
		return c.Send(payload)
	}
}

func (h *handler) updateProvider() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToUpdateProviderRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		res, err := h.crud.UpdateProvider(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromUpdateProviderResponseEntityToHTTPResponse(res)

		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) deleteProvider() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToDeleteProviderRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteProvider(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
