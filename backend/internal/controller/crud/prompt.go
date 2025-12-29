package crud

import (
	"context"
	"errors"
	"fmt"
	"time"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	modelmapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/model"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PromptController interface {
	CreatePrompt(ctx context.Context, req entity.CreatePromptRequest) (*entity.CreatePromptResponse, error)
	GetPrompt(ctx context.Context, req entity.GetPromptRequest) (*entity.GetPromptResponse, error)
	ListPrompts(ctx context.Context, req entity.ListPromptsRequest) (*entity.ListPromptsResponse, error)
	UpdatePrompt(ctx context.Context, req entity.UpdatePromptRequest) error
	DeletePrompt(ctx context.Context, req entity.DeletePromptRequest) error
}

const (
	_validationAttrPromptNameMaxLength        = 128
	_validationAttrPromptDescriptionMaxLength = 4096
)

func (c *controller) CreatePrompt(ctx context.Context, req entity.CreatePromptRequest) (*entity.CreatePromptResponse, error) {
	if err := c.validateCreatePromptRequest(req); err != nil {
		return nil, err
	}

	p := req.Prompt

	now := time.Now().UTC()
	prompt := model.Prompt{
		ID:          c.idgen.Next(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Name:        p.Name,
		Description: p.Description,
		Arguments:   p.Arguments,
		Messages:    p.Messages,
	}

	err := c.storage.CreatePrompt(ctx, prompt)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create server prompt",
			Data: map[string]any{
				"reason": err.Error(),
				"name":   p.Name,
			},
		}
	}

	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType: entity.ObjectTypePrompt,
		EventType:  entity.ObjectEventTypeCreate,
		ResoureID:  prompt.ID,
	})

	return &entity.CreatePromptResponse{Prompt: modelmapper.FromPromptModelToPromptEntity(prompt)}, nil
}

func (c *controller) GetPrompt(ctx context.Context, req entity.GetPromptRequest) (*entity.GetPromptResponse, error) {
	prompt, err := c.storage.GetPrompt(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeNotFound,
				Message: "server prompt not found",
				Data:    map[string]any{"id": req.ID},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get server prompt",
			Data:    map[string]any{"id": req.ID, "reason": err.Error()},
		}
	}

	return &entity.GetPromptResponse{
		Prompt: modelmapper.FromPromptModelToPromptEntity(*prompt),
	}, nil
}

func (c *controller) ListPrompts(ctx context.Context, req entity.ListPromptsRequest) (*entity.ListPromptsResponse, error) {
	prompts, err := c.storage.ListPrompts(ctx, nil)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to list server prompts",
			Data:    map[string]any{"reason": err.Error()},
		}
	}

	return &entity.ListPromptsResponse{
		Prompts: modelmapper.FromPromptModelsToPromptEntities(prompts),
	}, nil
}

func (c *controller) UpdatePrompt(ctx context.Context, req entity.UpdatePromptRequest) error {
	defer c.cache.Evict(ctx, entity.ObjectTypePrompt, req.Prompt.ID)

	if err := c.validateUpdatePromptRequest(req); err != nil {
		return err
	}

	p := req.Prompt

	attrs := make(map[model.PromptAttribute]any)
	if p.Name != "" {
		attrs[model.PromptAttributeName] = p.Name
	}
	if p.Description != "" {
		attrs[model.PromptAttributeDescription] = p.Description
	}
	if p.Arguments != nil {
		attrs[model.PromptAttributeArguments] = p.Arguments
	}
	if p.Messages != nil {
		attrs[model.PromptAttributeMessages] = p.Messages
	}

	if len(attrs) == 0 {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "no update fields provided",
			Data:    map[string]any{"id": p.ID},
		}
	}

	err := c.storage.UpdatePrompt(ctx, p.ID, attrs)
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update server prompt",
			Data:    map[string]any{"id": p.ID, "reason": err.Error()},
		}
	}

	serverIDs, err := c.storage.ListServerIDsByPromptID(ctx, req.Prompt.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by prompt id")
	}
	c.cache.Evict(ctx, entity.ObjectTypePrompt, req.Prompt.ID)
	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypePrompt,
			EventType:       entity.ObjectEventTypeDelete,
			ResoureID:       req.Prompt.ID,
			ResourceOwnerID: id,
		})
	}

	return nil
}

func (c *controller) DeletePrompt(ctx context.Context, req entity.DeletePromptRequest) error {
	serverIDs, err := c.storage.ListServerIDsByPromptID(ctx, req.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by prompt id")
	}

	err = c.storage.DeletePrompt(ctx, req.ID)
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete server prompt",
			Data:    map[string]any{"id": req.ID, "reason": err.Error()},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypePrompt, req.ID)
	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypePrompt,
			EventType:       entity.ObjectEventTypeDelete,
			ResoureID:       req.ID,
			ResourceOwnerID: id,
		})
	}

	return nil
}

func (c *controller) validateCreatePromptRequest(req entity.CreatePromptRequest) error {
	p := req.Prompt
	if p.Name == "" {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "prompt name is required"}
	}
	if len(p.Name) > _validationAttrPromptNameMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("prompt name exceeds maximum length of %d", _validationAttrPromptNameMaxLength)}
	}
	if len(p.Description) > _validationAttrPromptDescriptionMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("prompt description exceeds maximum length of %d", _validationAttrPromptDescriptionMaxLength)}
	}
	if p.Messages == nil {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "prompt messages are required"}
	}
	return nil
}

func (c *controller) validateUpdatePromptRequest(req entity.UpdatePromptRequest) error {
	p := req.Prompt
	if p.ID == 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "prompt ID is required"}
	}
	if len(p.Name) > _validationAttrPromptNameMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("prompt name exceeds maximum length of %d", _validationAttrPromptNameMaxLength)}
	}
	if len(p.Description) > _validationAttrPromptDescriptionMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("prompt description exceeds maximum length of %d", _validationAttrPromptDescriptionMaxLength)}
	}
	return nil
}
