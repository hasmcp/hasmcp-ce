package crud

import (
	"context"
	"errors"
	"fmt"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"gorm.io/gorm"
)

type ServerToolController interface {
	CreateServerTool(ctx context.Context, req entity.CreateServerToolRequest) (*entity.CreateServerToolResponse, error)
	DeleteServerTool(ctx context.Context, req entity.DeleteServerToolRequest) error
	ListServerTools(ctx context.Context, req entity.ListServerToolsRequest) (*entity.ListServerToolsResponse, error)
}

func (c *controller) CreateServerTool(
	ctx context.Context, req entity.CreateServerToolRequest) (*entity.CreateServerToolResponse, error) {
	if err := c.validateCreateServerToolRequest(req); err != nil {
		return nil, err
	}

	e := req.Tool

	// Check if server exists
	_, err := c.GetServer(ctx, entity.GetServerRequest{
		ID: e.ServerID,
	})
	if err != nil {
		return nil, err
	}

	tool, err := c.storage.GetProviderTool(ctx, req.Tool.ToolID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeConflict,
			Message: "tool does not exist",
			Data: map[string]any{
				"reason":   err.Error(),
				"serverID": e.ServerID,
				"toolID":   e.ToolID,
			},
		}
	}

	dt := model.ServerTool{
		ServerID:   e.ServerID,
		ProviderID: tool.ProviderID,
		ToolID:     e.ToolID,
	}
	err = c.storage.AddToolToServer(ctx, dt)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeConflict,
				Message: "server tool already exists",
				Data: map[string]any{
					"reason":   err.Error(),
					"serverID": e.ServerID,
					"toolID":   e.ToolID,
				},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create server tool",
			Data: map[string]any{
				"reason":   err.Error(),
				"serverID": e.ServerID,
				"toolID":   e.ToolID,
			},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeServer, req.Tool.ServerID)
	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerTool,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       e.ToolID,
		ResourceOwnerID: e.ServerID,
	})

	return &entity.CreateServerToolResponse{}, nil
}

func (c *controller) DeleteServerTool(ctx context.Context, req entity.DeleteServerToolRequest) error {
	if err := c.validateDeleteServerToolRequest(req); err != nil {
		return err
	}

	// Check if server exists
	_, err := c.GetServer(ctx, entity.GetServerRequest{
		ID: req.Tool.ServerID,
	})
	if err != nil {
		return err
	}

	err = c.storage.RemoveServerTool(ctx, model.ServerTool{
		ServerID: req.Tool.ServerID,
		ToolID:   req.Tool.ToolID,
	})
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete server tool",
			Data: map[string]any{
				"reason":   err.Error(),
				"serverID": req.Tool.ServerID,
				"toolID":   req.Tool.ToolID,
			},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeServer, req.Tool.ServerID)
	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerTool,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       req.Tool.ToolID,
		ResourceOwnerID: req.Tool.ServerID,
	})

	return nil
}

func (c *controller) ListServerTools(ctx context.Context, req entity.ListServerToolsRequest) (*entity.ListServerToolsResponse, error) {
	if req.ServerID <= 0 {
		return nil, fmt.Errorf("invalid server ID")
	}

	dts, err := c.storage.ListServerTools(ctx, req.ServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list server tokens: %w", err)
	}

	entities := make([]entity.ServerTool, 0, len(dts))
	for _, dt := range dts {
		entities = append(entities, entity.ServerTool{
			ProviderID: dt.ProviderID,
			ToolID:     dt.ToolID,
			ServerID:   dt.ServerID,
		})
	}

	return &entity.ListServerToolsResponse{
		Tools: entities,
	}, nil
}

func (c *controller) validateCreateServerToolRequest(req entity.CreateServerToolRequest) error {
	e := req.Tool
	if e.ServerID <= 0 {
		return fmt.Errorf("server ID must be greater than 0")
	}
	if e.ToolID <= 0 {
		return fmt.Errorf("tool ID must be greater than 0")
	}
	return nil
}

func (c *controller) validateDeleteServerToolRequest(req entity.DeleteServerToolRequest) error {
	e := req.Tool
	if e.ServerID <= 0 {
		return fmt.Errorf("server ID must be greater than 0")
	}
	if e.ToolID <= 0 {
		return fmt.Errorf("tool ID must be greater than 0")
	}
	return nil
}
