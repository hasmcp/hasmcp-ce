package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathPrompts      = "/prompts"
	_routePathCreatePrompt = _routePathPrompts
	_routePathListPrompts  = _routePathPrompts
	_routePathGetPrompt    = _routePathPrompts + "/:id"
	_routePathPatchPrompt  = _routePathPrompts + "/:id"
	_routePathDeletePrompt = _routePathPrompts + "/:id"
)

func (h *handler) registerPromptRoutes() error {
	h.router.Post(_routePathCreatePrompt, h.createPrompt())
	h.router.Get(_routePathListPrompts, h.listPrompts())
	h.router.Get(_routePathGetPrompt, h.getPrompt())
	h.router.Patch(_routePathPatchPrompt, h.updatePrompt())
	h.router.Delete(_routePathDeletePrompt, h.deletePrompt())

	return nil
}

func (h *handler) createPrompt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToCreatePromptRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreatePrompt(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreatePromptResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listPrompts() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToListPromptsRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.ListPrompts(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListPromptsResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) getPrompt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToGetPromptRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.GetPrompt(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromGetPromptResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) updatePrompt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToUpdatePromptRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.UpdatePrompt(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}

func (h *handler) deletePrompt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToDeletePromptRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeletePrompt(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
