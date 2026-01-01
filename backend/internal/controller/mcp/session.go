package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/pubsub"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

// DeleteSession serves to a regular tool
func (c *controller) DeleteSession(ctx context.Context, req DeleteSessionRequest) error {
	if _, ok := req.Permissions[ScopeSessionDelete]; !ok {
		return erre.Error{
			Code:    403,
			Message: "Insufficient permissions to delete the session",
		}
	}
	// session
	sessionRes, err := c.jwt.VerifySession(ctx, jwt.AuthParams{
		AccessToken: []byte(req.McpSessionID),
	})
	if err != nil {
		return erre.Error{
			Code:    404,
			Message: "Please initialize a new MCP session to make request",
			Data: map[string]any{
				"reason":         err.Error(),
				"Mcp-Session-Id": req.McpSessionID,
			},
		}
	}

	if sessionRes.ServerID != req.ServerID {
		return erre.Error{
			Code:    403,
			Message: "Please initialize a new MCP session to make request",
			Data: map[string]any{
				"reason":         "server-session mismatch!",
				"Mcp-Session-Id": req.McpSessionID,
			},
		}
	}

	sessionID := sessionRes.SessionID
	session, err := c.getSession(req.ServerID, sessionID)
	if err != nil {
		return nil
	}

	s, err := c.getServer(req.ServerID)
	if err != nil {
		return err
	}
	s.sessions.Delete(sessionID)

	// delegate to hasmcp/pubsub
	err = c.pubsub.Delete(ctx, pubsub.DeletePubSubRequest{
		ID: session.pubsubID,
	})
	if err != nil {
		return erre.Error{
			Code:    500, // This is not a JSONRPC call, but a regular http call, use http status codes
			Message: "Failed to delete PubSub",
			Data: map[string]any{
				"details": err.Error(),
			},
		}
	}

	session = nil

	return nil
}

// CallSession allows mcp client to execute protocol commands
func (c *controller) CallSession(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	if _, ok := req.Permissions[ScopeSessionCall]; !ok {
		return nil, erre.Error{
			Code:    403,
			Message: "Insufficient permissions to call the session",
		}
	}

	var err error
	var res *CallSessionResponse
	sessionInfo := req.McpSessionID

	if req.Request.Method != string(MethodInitialize) {
		sessionRes, err := c.jwt.VerifySession(ctx, jwt.AuthParams{
			AccessToken: []byte(sessionInfo),
		})
		if err != nil {
			return nil, jsonrpc.Error{
				Code:    jsonrpc.ErrCodeMethodNotFound,
				Message: "Please initialize a new MCP session to make request",
				Data: map[string]any{
					"reason":         err.Error(),
					"Mcp-Session-Id": req.McpSessionID,
				},
			}
		}

		if sessionRes.ServerID != req.ServerID {
			return nil, jsonrpc.Error{
				Code:    jsonrpc.ErrCodeMethodNotFound,
				Message: "Please initialize a new MCP session to make request",
				Data: map[string]any{
					"reason":         "server-session mismatch!",
					"Mcp-Session-Id": req.McpSessionID,
				},
			}
		}

		sessionInfo = fmt.Sprintf(
			"%s.%s/%s",
			monoflake.ID(sessionRes.SessionID).String(),
			sessionRes.InitializeParams.ClientInfo.Name,
			sessionRes.InitializeParams.ProtocolVersion,
		)
	}

	eventType := fmt.Sprintf("%s.%s", sessionInfo, req.Request.Method)

	_, err = c.pubsub.Publish(ctx, pubsub.PublishRequest{
		PubSubID: req.ServerID,
		Event: &event{
			Type: "« " + eventType,
			Data: req.Request.Params,
		},
	})
	if err != nil {
		if e, ok := err.(erre.Error); ok && e.Code == 404 {
			_, _ = c.pubsub.Create(ctx, pubsub.CreatePubSubRequest{
				ID: req.ServerID,
			})
			_, err = c.pubsub.Publish(ctx, pubsub.PublishRequest{
				PubSubID: req.ServerID,
				Event: &event{
					Type: "« " + eventType,
					Data: req.Request.Params,
				},
			})
			if err != nil {
				zlog.Error().Err(err).Msg("failed to publish after create attempt")
			}
		} else {
			zlog.Error().Err(err).Msg("failed to publish")
		}
	}

	// call desired method
	switch Method(req.Request.Method) {
	case MethodPing:
		res, err = c.CallPing(ctx, req) // implemented
	case MethodToolsList:
		res, err = c.CallToolsList(ctx, req) // implemented
	case MethodToolsCall:
		res, err = c.CallToolsCall(ctx, req) // implemented
	case MethodInitialize:
		res, err = c.CallInitialize(ctx, req) // implemented
	case MethodPromptsList:
		res, err = c.CallPromptsList(ctx, req) // implemented
	case MethodPromptsGet:
		res, err = c.CallPromptsGet(ctx, req) // implemented
	case MethodResourcesList:
		res, err = c.CallResourcesList(ctx, req) // implemented
	case MethodResourcesRead:
		res, err = c.CallResourcesRead(ctx, req) // implemented
	case MethodResourcesSubscribe:
		res, err = c.CallResourcesSubscribe(ctx, req) // not implemented
	case MethodResourcesTemplatesList:
		res, err = c.CallResourcesTemplatesList(ctx, req) // not implemented
	case MethodNotificationInitialize:
		if _, ok := req.Permissions[ScopeSessionCreate]; !ok {
			return nil, erre.Error{
				Code:    403,
				Message: "Insufficient permissions to create the session",
			}
		}
		res, err = c.CallNotificationsInitialized(ctx, req) // implemented and update session as client initialized
	case MethodNotificationRootsListChanged: // implemented but not functional
		res, err = c.CallNotificationsRootsListChanged(ctx, req)
	default:
		zlog.Warn().Str("method", req.Request.Method).Msg("RPC method not found!")
		err = jsonrpc.Error{
			Code:    jsonrpc.ErrCodeMethodNotFound,
			Message: "Method not found",
			Data: map[string]any{
				"method": req.Request.Method,
			},
		}
	}

	if err != nil {
		errStr, _ := json.Marshal(err)
		_, _ = c.pubsub.Publish(ctx, pubsub.PublishRequest{
			PubSubID: req.ServerID,
			Event: &event{
				Type: "» " + eventType,
				Data: errStr,
			},
		})
		zlog.Info().Err(err).Object("req", req).Msg(_logPrefix + "method call failed")
		return nil, err
	}

	var result json.RawMessage
	if res.Result != nil {
		result = res.Result.Result
	}
	_, err = c.pubsub.Publish(ctx, pubsub.PublishRequest{
		PubSubID: req.ServerID,
		Event: &event{
			Type: "» " + eventType,
			Data: result,
		},
	})
	if err != nil {
		zlog.Error().Err(err).Msg("failed to publish")
	}

	return res, nil
}

func (c *controller) SubscribeSession(ctx context.Context, req SubscribeSessionRequest) (*SubscribeSessionResponse, error) {
	if _, ok := req.Permissions[ScopeSessionStream]; !ok {
		return nil, erre.Error{
			Code:    403,
			Message: "Insufficient permissions to stream the session",
		}
	}

	// session
	sessionRes, err := c.jwt.VerifySession(ctx, jwt.AuthParams{
		AccessToken: []byte(req.McpSessionID),
	})
	if err != nil {
		return nil, erre.Error{
			Code:    404,
			Message: "Please initialize a new MCP session to make request",
			Data: map[string]any{
				"reason":         err.Error(),
				"Mcp-Session-Id": req.McpSessionID,
			},
		}
	}

	if sessionRes.ServerID != req.ServerID {
		return nil, erre.Error{
			Code:    403,
			Message: "Please initialize a new MCP session to make request",
			Data: map[string]any{
				"reason":         "server-session mismatch!",
				"Mcp-Session-Id": req.McpSessionID,
			},
		}
	}

	sessionID := sessionRes.SessionID
	_, err = c.getSession(req.ServerID, sessionID)
	if err != nil {
		// Re-add the session in here to recover a broken session due to server restart
		server, err := c.getServer(req.ServerID)
		if err != nil {
			return nil, err
		}
		server.sessions.Store(sessionID, &serverSession{
			pubsubID:         sessionRes.SessionID,
			initializeParams: sessionRes.InitializeParams,
		})

		// upsert pubsub
		_, _ = c.pubsub.Create(ctx, pubsub.CreatePubSubRequest{
			ID: sessionID,
		})
	}

	res, err := c.pubsub.Subscribe(ctx, pubsub.SubscribeRequest{
		PubSubID: sessionID,
	})
	if err != nil {
		return nil, err
	}

	return &SubscribeSessionResponse{
		Events:         res.Events,
		SubscriptionID: res.ID,
	}, nil
}

func (c *controller) UnsubscribeSession(ctx context.Context, req UnsubscribeSessionRequest) error {
	if _, ok := req.Permissions[ScopeSessionStream]; !ok {
		return erre.Error{
			Code:    403,
			Message: "Insufficient permissions to unsubscribe",
		}
	}

	// session
	sessionRes, err := c.jwt.VerifySession(ctx, jwt.AuthParams{
		AccessToken: []byte(req.McpSessionID),
	})
	if err != nil {
		return erre.Error{
			Code:    404,
			Message: "Please initialize a new MCP session to make request",
			Data: map[string]any{
				"reason":         err.Error(),
				"Mcp-Session-Id": req.McpSessionID,
			},
		}
	}

	return c.pubsub.Unsubscribe(ctx, pubsub.UnsubscribeRequest{
		PubSubID: sessionRes.SessionID,
		ID:       req.SubscriptionID,
	})
}

func (c *controller) sendSessionNotification(ctx context.Context, req CallSessionRequest) error {
	sessionInfo := req.McpSessionID
	session, err := c.getSession(req.ServerID, int64(monoflake.IDFromBase62(req.McpSessionID)))
	if err == nil {
		sessionInfo += "." + session.initializeParams.ClientInfo.Name
	}

	eventType := fmt.Sprintf("%s.%s.%s", sessionInfo, _serverProtocolVersion, req.Request.Method)
	_, _ = c.pubsub.Publish(ctx, pubsub.PublishRequest{
		PubSubID: req.ServerID,
		Event: &event{
			Type: "» " + eventType,
			Data: req.Request.Params,
		},
	})

	switch req.Request.Method {
	case MethodNotificationToolsListChanged: // server to client
		_, err = c.CallNotificationsToolsListChanged(ctx, req)
	case MethodNotificationPromptsListChanged: // server to client
		_, err = c.CallNotificationsPromptsListChanged(ctx, req)
	case MethodNotificationResourcesListChanged: // server to client
		_, err = c.CallNotificationsResourceListChanged(ctx, req)
	}

	if err != nil {
		_, _ = c.pubsub.Publish(ctx, pubsub.PublishRequest{
			PubSubID: req.ServerID,
			Event: &event{
				Type: "i " + eventType,
				Data: []byte(err.Error()),
			},
		})
	}

	return err
}
