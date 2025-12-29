package api

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp"
	jwtc "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/mcp/middleware/cors"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/mcp/middleware/jwt"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/mcp/middleware/logger"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/mcp/middleware/ratelimit"
	apimapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/api"
	mapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/mcp"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/server"
	zlog "github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

type (
	Params struct {
		Config  config.Service
		Server  server.Service
		Mcp     mcp.Controller
		JWTAuth jwtc.AuthenticatorController
	}

	Handler interface {
	}

	handler struct {
		mcp    mcp.Controller
		router fiber.Router
	}
)

const (
	_logPrefix = "[mcp/h] "

	_routeBasePath = "/mcp"

	_headerContentType                     = "content-type"
	_headerContentTypeValueApplicationJSON = "application/json"
	_headerContentTypeValueTextEventStream = "text/event-stream"
)

func New(p Params) (Handler, error) {
	jwt, err := jwt.New(jwt.Params{
		Config:  p.Config,
		JWTAuth: p.JWTAuth,
	})
	if err != nil {
		return nil, err
	}
	logger, err := logger.New(logger.Params{
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}
	ratelimit, err := ratelimit.New(ratelimit.Params{
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}
	cors, err := cors.New(cors.Params{
		Config: p.Config,
	})

	if err != nil {
		return nil, err
	}

	router := p.Server.Group(_routeBasePath).Use(logger).Use(cors).Use(ratelimit).Use(jwt)
	h := &handler{
		router: router,
		mcp:    p.Mcp,
	}

	if err := h.registerMcpRoutes(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *handler) registerMcpRoutes() error {
	h.router.Get("/:id/logs", h.tail())
	h.router.Post("/:id", h.jsonRPC())
	h.router.Get("/:id", h.stream())
	h.router.Delete("/:id", h.delete())

	return nil
}

func (h *handler) jsonRPC() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(_headerContentType, _headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToMcpCallSessionRequest(c)
		if rq == nil {
			c.Status(400)
			return c.JSON(jsonrpc.ErrorResponse{
				JSONRpc: "2.0",
				Error: jsonrpc.Error{
					Code:    jsonrpc.ErrCodeInvalidJsonReceived,
					Message: "invalid jsonrpc 2.0 object received",
					Data: map[string]any{
						"payload": string(c.BodyRaw()),
					},
				},
				ID: nil,
			})
		}

		rs, err := h.mcp.CallSession(context.Background(), *rq)
		if err != nil {
			e, status := mapper.FromErrorToJsonRpcResponse(rq.Request.ID, err)
			c.Status(status)
			return c.Send(e)
		}

		payload := mapper.FromMcpCallSessionResponseToHTTPResponse(*rs)

		c.Set("mcp-session-id", rs.McpSessionID)
		c.Set("mcp-protocol-version", rq.McpProtocolVersion)
		c.Status(rs.HTTPStatusCode)
		return c.Send(payload)
	}
}

func (h *handler) stream() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rq := mapper.FromHTTPRequestToMcpSubscribeSessionRequest(c)
		rs, err := h.mcp.SubscribeSession(context.Background(), rq)
		if err != nil {
			c.Set(_headerContentType, _headerContentTypeValueApplicationJSON)
			e, status := apimapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}
		origin := string(c.Get("origin"))
		if origin == "" {
			origin = "*"
		}

		zlog.Info().Str("sessionID", rq.McpSessionID).Msg(_logPrefix + "streaming is initialized")

		c.Set(_headerContentType, _headerContentTypeValueTextEventStream)
		c.Context().SetConnectionClose()
		c.Set("cache-control", "no-cache")
		c.Set("connection", "keep-alive")
		c.Set("transfer-encoding", "chunked")
		c.Set("access-control-allow-origin", origin)
		c.Set("access-control-allow-headers", "cache-control")
		c.Set("access-control-allow-credentials", "true")

		ctx := c.Status(fiber.StatusOK).Context()
		ctx.SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			var id string
			var event string
			var data []byte
			var validType bool
			var ok bool
			var e any
			var sse mcp.SSE
			freshCtx := context.Background()

			zlog.Info().
				Int64("serverID", rq.ServerID).
				Int64("subscriptionID", rs.SubscriptionID).
				Str("sessionID", rq.McpSessionID).
				Msg(_logPrefix + "sse conn opened by user")
			ticker := time.NewTicker(time.Second * 3)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					zlog.Info().
						Int64("serverID", rq.ServerID).
						Int64("subscriptionID", rs.SubscriptionID).
						Str("sessionID", rq.McpSessionID).
						Msg(_logPrefix + "sse conn closed by user")
					err := h.mcp.UnsubscribeSession(freshCtx, mcp.UnsubscribeSessionRequest{
						McpSessionID:   rq.McpSessionID,
						SubscriptionID: rs.SubscriptionID,
						Permissions:    rq.Permissions,
					})
					if err != nil {
						zlog.Warn().Err(err).
							Int64("serverID", rq.ServerID).
							Int64("subscriptionID", rs.SubscriptionID).
							Str("sessionID", rq.McpSessionID).
							Msg(_logPrefix + "failed to unsubscribe on user disconnect")
					}
					return
				case <-ticker.C:
					fmt.Fprintf(w, ": {\"status\": \"tick\"}\n\n")
					if err := w.Flush(); err != nil {
						zlog.Warn().Err(err).
							Int64("serverID", rq.ServerID).
							Int64("subscriptionID", rs.SubscriptionID).
							Str("sessionID", rq.McpSessionID).
							Msg(_logPrefix + "failed to flush on tick")
						err := h.mcp.UnsubscribeSession(freshCtx, mcp.UnsubscribeSessionRequest{
							McpSessionID:   rq.McpSessionID,
							SubscriptionID: rs.SubscriptionID,
							Permissions:    rq.Permissions,
						})
						if err != nil {
							zlog.Warn().Err(err).
								Int64("serverID", rq.ServerID).
								Int64("subscriptionID", rs.SubscriptionID).
								Str("sessionID", rq.McpSessionID).
								Msg(_logPrefix + "failed to unsubscribe on tick flush failure")
						}
						return
					}
				case e, ok = <-rs.Events:
					if !ok {
						// channel closed
						return
					}

					sse, ok = e.(mcp.SSE)
					if !ok {
						zlog.Warn().Msg("unknown event type")
						continue
					}

					id = sse.GetID()
					if len(id) > 0 {
						fmt.Fprintf(w, "id: %s\n", id)
					}

					event = sse.GetType()
					if len(event) > 0 {
						fmt.Fprintf(w, "event: %s\n", event)
					}

					data, validType = sse.GetData().([]byte)
					if validType {
						fmt.Fprintf(w, "data: %s\n\n", string(data))
						data = nil
					}

					if err := w.Flush(); err != nil {
						zlog.Error().Err(err).
							Int64("serverID", rq.ServerID).
							Int64("subscriptionID", rs.SubscriptionID).
							Str("sessionID", rq.McpSessionID).
							Msg("failed to flush on event")
						err := h.mcp.UnsubscribeSession(freshCtx, mcp.UnsubscribeSessionRequest{
							McpSessionID:   rq.McpSessionID,
							SubscriptionID: rs.SubscriptionID,
							Permissions:    rq.Permissions,
						})
						if err != nil {
							zlog.Warn().Err(err).
								Int64("serverID", rq.ServerID).
								Int64("subscriptionID", rs.SubscriptionID).
								Str("sessionID", rq.McpSessionID).
								Msg("failed to unsubscribe on message flush failure")
						}
						return
					}
				}
			}
		}))

		return nil
	}
}

func (h *handler) tail() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rq := mapper.FromHTTPRequestToMcpStartTailIORequest(c)
		rs, err := h.mcp.StartTailIO(context.Background(), rq)
		if err != nil {
			c.Set(_headerContentType, _headerContentTypeValueApplicationJSON)
			e, status := apimapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}
		origin := string(c.Get("origin"))
		if origin == "" {
			origin = "*"
		}

		zlog.Info().Int64("serverID", rq.ServerID).Msg("tailing is initialized")

		c.Set(_headerContentType, _headerContentTypeValueTextEventStream)
		c.Context().SetConnectionClose()
		c.Set("cache-control", "no-cache")
		c.Set("connection", "keep-alive")
		c.Set("transfer-encoding", "chunked")
		c.Set("access-control-allow-origin", origin)
		c.Set("access-control-allow-headers", "cache-control")
		c.Set("access-control-allow-credentials", "true")

		ctx := c.Status(fiber.StatusOK).Context()
		ctx.SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			var id string
			var event string
			var data []byte
			var validType bool
			var ok bool
			var e any
			var sse mcp.SSE
			freshCtx := context.Background()

			zlog.Info().Int64("serverID", rq.ServerID).
				Int64("subscriptionID", rs.SubscriptionID).
				Int64("serverID", rq.ServerID).
				Msg("sse conn opened by user")
			ticker := time.NewTicker(time.Second * 3)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					zlog.Info().
						Int64("serverID", rq.ServerID).
						Int64("subscriptionID", rs.SubscriptionID).
						Msg(_logPrefix + "sse conn closed by user")
					err := h.mcp.StopTailIO(freshCtx, mcp.StopTailIORequest{
						ServerID:       rq.ServerID,
						SubscriptionID: rs.SubscriptionID,
						Permissions:    rq.Permissions,
					})
					if err != nil {
						zlog.Warn().Err(err).
							Int64("serverID", rq.ServerID).
							Int64("subscriptionID", rs.SubscriptionID).
							Msg(_logPrefix + "failed to unsubscribe on user disconnect")
					}
					return
				case <-ticker.C:
					fmt.Fprintf(w, ": {\"status\": \"tick\"}\n\n")
					if err := w.Flush(); err != nil {
						zlog.Warn().Err(err).
							Int64("serverID", rq.ServerID).
							Int64("subscriptionID", rs.SubscriptionID).
							Msg(_logPrefix + "failed to flush on tick")
						err := h.mcp.StopTailIO(freshCtx, mcp.StopTailIORequest{
							ServerID:       rq.ServerID,
							SubscriptionID: rs.SubscriptionID,
							Permissions:    rq.Permissions,
						})
						if err != nil {
							zlog.Warn().Err(err).
								Int64("serverID", rq.ServerID).
								Int64("subscriptionID", rs.SubscriptionID).
								Msg(_logPrefix + "failed to unsubscribe on tick flush failure")
						}
						return
					}
				case e, ok = <-rs.Events:
					if !ok {
						// channel closed
						return
					}

					sse, ok = e.(mcp.SSE)
					if !ok {
						zlog.Warn().Msg("unknown event type")
						continue
					}

					id = sse.GetID()
					if len(id) > 0 {
						fmt.Fprintf(w, "id: %s\n", id)
					}

					event = sse.GetType()
					if len(event) > 0 {
						fmt.Fprintf(w, "event: %s\n", event)
					}

					data, validType = sse.GetData().([]byte)
					if validType {
						fmt.Fprintf(w, "data: %s\n\n", string(data))
						data = nil
					}

					if err := w.Flush(); err != nil {
						zlog.Error().Err(err).
							Int64("serverID", rq.ServerID).
							Int64("subscriptionID", rs.SubscriptionID).
							Msg("failed to flush on event")
						err := h.mcp.StopTailIO(freshCtx, mcp.StopTailIORequest{
							ServerID:       rq.ServerID,
							SubscriptionID: rs.SubscriptionID,
							Permissions:    rq.Permissions,
						})
						if err != nil {
							zlog.Warn().Err(err).
								Int64("serverID", rq.ServerID).
								Int64("subscriptionID", rs.SubscriptionID).
								Msg("failed to unsubscribe on message flush failure")
						}
						return
					}
				}
			}
		}))

		return nil
	}
}

func (h *handler) delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(_headerContentType, _headerContentTypeValueApplicationJSON)
		rq := mapper.FromHTTPRequestToMcpDeleteSessionRequest(c)
		err := h.mcp.DeleteSession(context.Background(), rq)
		if err != nil {
			e, status := apimapper.FromErrorToHTTPResponse(err)
			c.Status(status)
			return c.Send(e)
		}

		c.Status(http.StatusNoContent)
		c.Send([]byte{})
		return nil
	}
}
