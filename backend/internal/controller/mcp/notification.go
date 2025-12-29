package mcp

import (
	"context"

	"github.com/hasmcp/hasmcp-ce/backend/internal/service/pubsub"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

const (
	// MethodNotificationResourcesListChanged notifies when the list of available resources changes.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/resources#list-changed-notification
	// Server to Client
	MethodNotificationResourcesListChanged = "notifications/resources/list_changed"

	// MethodNotificationResourceUpdated notifies when the available resource is updated
	// Server to Client
	MethodNotificationResourceUpdated = "notifications/resources/updated"

	// MethodNotificationPromptsListChanged notifies when the list of available prompt templates changes.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/prompts#list-changed-notification
	// Server to Client
	MethodNotificationPromptsListChanged = "notifications/prompts/list_changed"

	// MethodNotificationToolsListChanged notifies when the list of available tools changes.
	// https://modelcontextprotocol.io/specification/2025-06-18/schema#notifications%2Ftools%2Flist-changed
	// Server to Client
	MethodNotificationToolsListChanged = "notifications/tools/list_changed"

	/* Below methods are client to server notifications */

	// MethodNotificationInitialize notifies when the the initialization completes
	// hhttps://modelcontextprotocol.io/specification/2024-11-05/basic/lifecycle#initialization
	// Client to Server
	MethodNotificationInitialize = "notifications/initialized"

	// MethodNotificationRootsListChanged when roots change, clients that support listChanged MUST send a notification
	// https://modelcontextprotocol.io/specification/2025-06-18/client/roots
	// Client to Server
	MethodNotificationRootsListChanged = "notificartions/roots/list_changed"
)

var (
	_notificationPayloadToolsListChanged     = []byte(`{"jsonrpc": "2.0", "method": "notifications/tools/list_changed"}`)
	_notificationPayloadPromptsListChanged   = []byte(`{"jsonrpc": "2.0", "method": "notifications/prompts/list_changed"}`)
	_notificationPayloadResourcesListChanged = []byte(`{"jsonrpc": "2.0", "method": "notifications/resources/list_changed"}`)
)

func (c *controller) CallNotificationsInitialized(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	return &CallSessionResponse{
		HTTPStatusCode:     202,
		McpSessionID:       req.McpSessionID,
		McpProtocolVersion: req.McpProtocolVersion,
		Result:             nil,
	}, nil
}

func (c *controller) CallNotificationsRootsListChanged(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	return &CallSessionResponse{
		HTTPStatusCode:     202,
		McpSessionID:       req.McpSessionID,
		McpProtocolVersion: req.McpProtocolVersion,
		Result:             nil,
	}, nil
}

func (c *controller) CallNotificationsResourceListChanged(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	_, err := c.pubsub.Publish(ctx, pubsub.PublishRequest{
		PubSubID: monoflake.IDFromBase62(req.McpSessionID).Int64(),
		Event: &event{
			Data: _notificationPayloadResourcesListChanged,
		},
	})
	if err != nil {
		zlog.Error().Err(err).Msg("failed to send notification")
		return nil, err
	}
	return &CallSessionResponse{}, nil
}

func (c *controller) CallNotificationsResourcesUpdated(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	return nil, ErrNotImplemented
}

func (c *controller) CallNotificationsPromptsListChanged(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	_, err := c.pubsub.Publish(ctx, pubsub.PublishRequest{
		PubSubID: monoflake.IDFromBase62(req.McpSessionID).Int64(),
		Event: &event{
			Data: _notificationPayloadPromptsListChanged,
		},
	})
	if err != nil {
		zlog.Error().Err(err).Msg("failed to send notification")
		return nil, err
	}
	return &CallSessionResponse{}, nil
}

func (c *controller) CallNotificationsToolsListChanged(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	_, err := c.pubsub.Publish(ctx, pubsub.PublishRequest{
		PubSubID: monoflake.IDFromBase62(req.McpSessionID).Int64(),
		Event: &event{
			Data: _notificationPayloadToolsListChanged,
		},
	})
	if err != nil {
		zlog.Error().Err(err).Msg("failed to send notification")
		return nil, err
	}
	return &CallSessionResponse{}, nil
}
