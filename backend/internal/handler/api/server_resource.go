package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathServerResourceAssociations      = _routePathServers + "/:id/resources"
	_routePathCreateServerResourceAssociation = _routePathServerResourceAssociations
	_routePathListServerResourceAssociations  = _routePathServerResourceAssociations
	_routePathDeleteServerResourceAssociation = _routePathServerResourceAssociations + "/:resourceID"
)

func (h *handler) registerServerResourceAssociationRoutes() error {
	h.router.Post(_routePathCreateServerResourceAssociation, h.createServerResourceAssociation())
	h.router.Get(_routePathListServerResourceAssociations, h.listServerResourceAssociations())
	h.router.Delete(_routePathDeleteServerResourceAssociation, h.deleteServerResourceAssociation())

	return nil
}

func (h *handler) createServerResourceAssociation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToCreateServerResourceAssociationRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateServerResource(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateServerResourceAssociationResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listServerResourceAssociations() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToListServerResourcesRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListServerResources(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListServerResourcesResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) deleteServerResourceAssociation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToDeleteServerResourceAssociationRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteServerResource(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
