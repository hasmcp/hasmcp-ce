package crud

import (
	"context"
	"errors"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"gorm.io/gorm"
)

type ServerPromptController interface {
	CreateServerPrompt(ctx context.Context, req entity.CreateServerPromptRequest) (*entity.CreateServerPromptResponse, error)
	DeleteServerPrompt(ctx context.Context, req entity.DeleteServerPromptRequest) error
	ListServerPrompts(ctx context.Context, req entity.ListServerPromptsRequest) (*entity.ListServerPromptsResponse, error)
}

func (c *controller) CreateServerPrompt(
	ctx context.Context, req entity.CreateServerPromptRequest) (*entity.CreateServerPromptResponse, error) {
	if err := c.validateCreateServerPromptRequest(req); err != nil {
		return nil, err
	}

	p := req.Prompt

	// Check if server exists
	if _, err := c.GetServer(ctx, entity.GetServerRequest{ID: p.ServerID}); err != nil {
		return nil, err
	}

	// Check if prompt exists
	if _, err := c.storage.GetPrompt(ctx, p.PromptID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeNotFound,
				Message: "prompt not found",
				Data:    map[string]any{"promptID": p.PromptID},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get prompt",
			Data:    map[string]any{"promptID": p.PromptID, "reason": err.Error()},
		}
	}

	prompt := model.ServerPrompt{
		ServerID: p.ServerID,
		PromptID: p.PromptID,
	}
	err := c.storage.AddPromptToServer(ctx, prompt)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeConflict,
				Message: "server prompt association already exists",
				Data:    map[string]any{"serverID": p.ServerID, "promptID": p.PromptID},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create server prompt association",
			Data:    map[string]any{"serverID": p.ServerID, "promptID": p.PromptID, "reason": err.Error()},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeServer, req.Prompt.ServerID)
	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerPrompt,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       p.PromptID,
		ResourceOwnerID: p.ServerID,
	})

	return &entity.CreateServerPromptResponse{}, nil
}

func (c *controller) DeleteServerPrompt(ctx context.Context, req entity.DeleteServerPromptRequest) error {
	if err := c.validateDeleteServerPromptRequest(req); err != nil {
		return err
	}
	err := c.storage.RemoveServerPrompt(ctx, model.ServerPrompt{
		ServerID: req.ServerID,
		PromptID: req.PromptID,
	})
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete server prompt association",
			Data:    map[string]any{"serverID": req.ServerID, "promptID": req.PromptID, "reason": err.Error()},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeServer, req.ServerID)
	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerPrompt,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       req.PromptID,
		ResourceOwnerID: req.ServerID,
	})

	return nil
}

func (c *controller) ListServerPrompts(ctx context.Context, req entity.ListServerPromptsRequest) (*entity.ListServerPromptsResponse, error) {
	if req.ServerID <= 0 {
		return nil, erre.Error{Code: erre.ErrorCodeBadRequest, Message: "invalid server ID"}
	}

	prompts, err := c.storage.ListServerPrompts(ctx, req.ServerID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to list server prompt associations",
			Data:    map[string]any{"serverID": req.ServerID, "reason": err.Error()},
		}
	}

	ps := make([]entity.ServerPrompt, len(prompts))
	for i, p := range prompts {
		ps[i] = entity.ServerPrompt{
			PromptID: p.PromptID,
			ServerID: p.ServerID,
		}
	}

	return &entity.ListServerPromptsResponse{
		Prompts: ps,
	}, nil
}

func (c *controller) validateCreateServerPromptRequest(req entity.CreateServerPromptRequest) error {
	p := req.Prompt
	if p.ServerID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "invalid server ID"}
	}
	if p.PromptID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "invalid prompt ID"}
	}
	return nil
}

func (c *controller) validateDeleteServerPromptRequest(req entity.DeleteServerPromptRequest) error {
	if req.ServerID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "invalid server ID"}
	}
	if req.PromptID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "invalid prompt ID"}
	}
	return nil
}
