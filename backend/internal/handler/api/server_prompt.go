package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathServerPromptAssociations      = _routePathServers + "/:id/prompts"
	_routePathCreateServerPromptAssociation = _routePathServerPromptAssociations
	_routePathListServerPromptAssociations  = _routePathServerPromptAssociations
	_routePathDeleteServerPromptAssociation = _routePathServerPromptAssociations + "/:promptID"
)

func (h *handler) registerServerPromptAssociationRoutes() error {
	h.router.Post(_routePathCreateServerPromptAssociation, h.createServerPromptAssociation())
	h.router.Get(_routePathListServerPromptAssociations, h.listServerPromptAssociations())
	h.router.Delete(_routePathDeleteServerPromptAssociation, h.deleteServerPromptAssociation())

	return nil
}

func (h *handler) createServerPromptAssociation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToCreateServerPromptAssociationRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateServerPrompt(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateServerPromptAssociationResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listServerPromptAssociations() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToListServerPromptsRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListServerPrompts(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListServerPromptsResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) deleteServerPromptAssociation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)

		rq := mapper.FromHTTPRequestToDeleteServerPromptAssociationRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteServerPrompt(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
