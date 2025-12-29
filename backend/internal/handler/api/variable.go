package api

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
)

const (
	_routePathVariables      = "/variables"
	_routePathCreateVariable = _routePathVariables
	_routePathListVariables  = _routePathVariables
	_routePathPatchVariable  = _routePathVariables + "/:id"
	_routePathDeleteVariable = _routePathVariables + "/:id"
)

func (h *handler) registerVariableRoutes() error {
	h.router.Post(_routePathCreateVariable, h.createVariable())
	h.router.Get(_routePathListVariables, h.listVariables())
	h.router.Patch(_routePathPatchVariable, h.updateVariable())
	h.router.Delete(_routePathDeleteVariable, h.deleteVariable())

	return nil
}

func (h *handler) createVariable() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToCreateVariableRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.CreateVariable(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromCreateVariableResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusCreated)
		return c.Send(payload)
	}
}

func (h *handler) listVariables() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rs, err := h.crud.ListVariables(context.Background())
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromListVariablesResponseEntityToHTTPResponse(rs)
		c.Status(http.StatusOK)
		return c.Send(payload)
	}
}

func (h *handler) updateVariable() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToUpdateVariableRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		rs, err := h.crud.UpdateVariable(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromUpdateVariableResponseEntityToHTTPResponse(rs)

		c.Status(http.StatusNoContent)
		return c.Send(payload)
	}
}

func (h *handler) deleteVariable() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(headerContentType, headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToDeleteVariableRequestEntity(c)
		if rq == nil {
			c.Status(http.StatusUnprocessableEntity)
			return c.Send(_invalidRequestPayloadHTTPError)
		}

		err := h.crud.DeleteVariable(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		return c.Send([]byte(""))
	}
}
