package mcp

import (
	"context"

	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/pubsub"
	"github.com/mustafaturan/monoflake"
)

type (
	event struct {
		ID   int64
		Type string
		Data []byte
	}

	UnsubscribeSessionRequest struct {
		McpSessionID   string
		SubscriptionID int64
		Permissions    map[string]struct{}
	}

	SubscribeSessionRequest struct {
		ServerID           int64
		McpSessionID       string
		McpProtocolVersion string
		LastEventID        string
		Permissions        map[string]struct{}
	}

	SubscribeSessionResponse struct {
		Events         chan any
		SubscriptionID int64
	}

	StartTailIORequest struct {
		ServerID    int64
		Permissions map[string]struct{}
	}

	StartTailIOResponse struct {
		Events         chan any
		SubscriptionID int64
	}

	StopTailIORequest struct {
		ServerID       int64
		SubscriptionID int64
		Permissions    map[string]struct{}
	}

	SSE interface {
		GetID() string
		GetType() string
		GetData() any
	}
)

func (e *event) GetID() string {
	if e.ID == 0 {
		return ""
	}
	return monoflake.ID(e.ID).String()
}

func (e *event) GetType() string {
	return e.Type
}

func (e *event) GetData() any {
	return e.Data
}

func (c *controller) StartTailIO(ctx context.Context, req StartTailIORequest) (*StartTailIOResponse, error) {
	if _, ok := req.Permissions[ScopeServerTail]; !ok {
		return nil, erre.Error{
			Code:    403,
			Message: "Insufficient scope",
			Data: map[string]any{
				"id": req.ServerID,
			},
		}
	}

	_, _ = c.pubsub.Create(ctx, pubsub.CreatePubSubRequest{
		ID: req.ServerID,
	})

	res, err := c.pubsub.Subscribe(ctx, pubsub.SubscribeRequest{
		PubSubID: req.ServerID,
	})
	if err != nil {
		return nil, err
	}

	return &StartTailIOResponse{
		Events:         res.Events,
		SubscriptionID: res.ID,
	}, nil
}

func (c *controller) StopTailIO(ctx context.Context, req StopTailIORequest) error {
	if _, ok := req.Permissions[ScopeServerTail]; !ok {
		return erre.Error{
			Code:    403,
			Message: "Insufficient scope",
			Data: map[string]any{
				"id": req.ServerID,
			},
		}
	}

	return c.pubsub.Unsubscribe(ctx, pubsub.UnsubscribeRequest{
		PubSubID: req.ServerID,
		ID:       req.SubscriptionID,
	})
}
