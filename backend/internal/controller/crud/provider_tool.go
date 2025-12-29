package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	modelmapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/model"
	"github.com/kaptinlin/jsonschema"
	zlog "github.com/rs/zerolog/log"
)

type ProviderToolController interface {
	GetProviderTool(ctx context.Context, req entity.GetProviderToolRequest) (*entity.GetProviderToolResponse, error)
	ListProviderTools(ctx context.Context, req entity.ListProviderToolsRequest) (*entity.ListProviderToolsResponse, error)
	CreateProviderTool(ctx context.Context, req entity.CreateProviderToolRequest) (*entity.CreateProviderToolResponse, error)
	UpdateProviderTool(ctx context.Context, req entity.UpdateProviderToolRequest) (*entity.UpdateProviderToolResponse, error)
	DeleteProviderTool(ctx context.Context, req entity.DeleteProviderToolRequest) error
}

const (
	_validationAttrProviderToolPathMinLength  = 1
	_validationAttrProviderToolPathMaxLength  = 128
	_validationAttrProviderToolDescMaxLength  = 4096
	_validationAttrProviderToolTitleMaxLength = 64
)

var (
	_regexValidationProviderToolName = regexp.MustCompile(`^[a-z][a-zA-Z0-9]{0,19}$`)
)

func (c *controller) CreateProviderTool(ctx context.Context, req entity.CreateProviderToolRequest) (*entity.CreateProviderToolResponse, error) {
	if err := c.validateCreateProviderToolRequest(req); err != nil {
		return nil, err
	}

	e := req.Tool

	headers, err := json.Marshal(e.Headers)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	tool := model.ProviderTool{
		ID:                  c.idgen.Next(),
		CreatedAt:           now,
		UpdatedAt:           now,
		ProviderID:          e.ProviderID,
		Method:              uint8(e.Method),
		Path:                e.Path,
		Name:                e.Name,
		Title:               e.Title,
		Description:         e.Description,
		PathArgsJSONSchema:  e.PathArgsJSONSchema,
		QueryArgsJSONSchema: e.QueryArgsJSONSchema,
		ReqBodyJSONSchema:   e.ReqBodyJSONSchema,
		ResBodyJSONSchema:   e.ResBodyJSONSchema,
		Headers:             headers,
		Oauth2Scopes:        strings.Join(e.Oauth2Scopes, ","),
	}

	// Init transaction
	ctx = c.storage.ContextWithTx(ctx)
	if err := c.storage.CreateProviderTool(ctx, tool); err != nil {
		_ = c.storage.TxRollback(ctx)
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create provider tool",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": e.ProviderID,
			},
		}
	}

	// Updates version!
	err = c.storage.UpdateProvider(ctx, e.ProviderID, nil)
	if err != nil {
		_ = c.storage.TxRollback(ctx)
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update provider version due to new tool creation",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": e.ProviderID,
			},
		}
	}

	err = c.storage.TxCommit(ctx)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "db transaction failed to provider tool creation",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": e.ProviderID,
			},
		}
	}

	return &entity.CreateProviderToolResponse{
		Tool: modelmapper.FromProviderToolModelToProviderToolEntity(tool),
	}, nil
}

func (c *controller) ListProviderTools(ctx context.Context, req entity.ListProviderToolsRequest) (*entity.ListProviderToolsResponse, error) {
	tools, err := c.storage.ListProviderTools(ctx, req.ProviderID, req.ToolIDs)
	if err != nil {
		return nil, err
	}

	return &entity.ListProviderToolsResponse{
		Tools: modelmapper.FromProviderToolModelsToProviderToolEntities(tools),
	}, nil
}

func (c *controller) GetProviderTool(ctx context.Context, req entity.GetProviderToolRequest) (*entity.GetProviderToolResponse, error) {
	e, err := c.storage.GetProviderTool(ctx, req.ToolID)
	if err != nil {
		return nil, err
	}

	return &entity.GetProviderToolResponse{
		Tool: modelmapper.FromProviderToolModelToProviderToolEntity(*e),
	}, nil
}

func (c *controller) UpdateProviderTool(ctx context.Context, req entity.UpdateProviderToolRequest) (*entity.UpdateProviderToolResponse, error) {
	if err := c.validateUpdateProviderToolRequest(req); err != nil {
		return nil, err
	}

	e := req.Tool

	attrs := make(map[model.ProviderToolAttribute]any)
	if e.Name != "" {
		attrs[model.ProviderToolAttributeName] = e.Name
	}
	if e.Title != "" {
		attrs[model.ProviderToolAttributeTitle] = e.Title
	}
	if e.Description != "" {
		attrs[model.ProviderToolAttributeDescription] = e.Description
	}
	if e.Description != "" {
		attrs[model.ProviderToolAttributeDescription] = e.Description
	}
	if len(e.PathArgsJSONSchema) > 0 {
		attrs[model.ProviderToolAttributePathArgsJSONSchema] = e.PathArgsJSONSchema
	}
	if len(e.QueryArgsJSONSchema) > 0 {
		attrs[model.ProviderToolAttributeQueryArgsJSONSchema] = e.QueryArgsJSONSchema
	}
	if len(e.ReqBodyJSONSchema) > 0 {
		attrs[model.ProviderToolAttributeReqBodyJSONSchema] = e.ReqBodyJSONSchema
	}
	if len(e.ResBodyJSONSchema) > 0 {
		attrs[model.ProviderToolAttributeResBodyJSONSchema] = e.ResBodyJSONSchema
	}
	if len(e.Oauth2Scopes) > 0 {
		attrs[model.ProviderToolAttributeOauth2Scopes] = strings.Join(e.Oauth2Scopes, ",")
	}
	if e.Headers != nil {
		headers, err := json.Marshal(e.Headers)
		if err != nil {
			return nil, err
		}
		attrs[model.ProviderToolAttributeHeaders] = headers
	}

	// Init transaction
	ctx = c.storage.ContextWithTx(ctx)
	err := c.storage.UpdateProviderTool(ctx, e.ID, attrs)
	if err != nil {
		_ = c.storage.TxRollback(ctx)
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update provider tool",
			Data: map[string]any{
				"reason": err.Error(),
				"toolID": e.ID,
			},
		}
	}

	// Updates version!
	err = c.storage.UpdateProvider(ctx, e.ProviderID, nil)
	if err != nil {
		_ = c.storage.TxRollback(ctx)
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update provider version due to tool changes",
			Data: map[string]any{
				"reason":     err.Error(),
				"toolID":     e.ID,
				"providerID": e.ProviderID,
			},
		}
	}

	err = c.storage.TxCommit(ctx)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "db transaction failed to update provider tool changes",
			Data: map[string]any{
				"reason":     err.Error(),
				"toolID":     e.ID,
				"providerID": e.ProviderID,
			},
		}
	}

	freshCtx := context.Background()

	toolRes, err := c.storage.GetProviderTool(freshCtx, e.ID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get the provider tool",
			Data: map[string]any{
				"reason":     err.Error(),
				"toolID":     e.ID,
				"providerID": e.ProviderID,
			},
		}
	}

	serverIDs, err := c.storage.ListServerIDsByToolID(freshCtx, req.Tool.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by tool id")
	}
	c.cache.Evict(freshCtx, entity.ObjectTypeProviderTool, req.Tool.ID)
	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(freshCtx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeProviderTool,
			EventType:       entity.ObjectEventTypeUpdate,
			ResoureID:       req.Tool.ID,
			ResourceOwnerID: id,
		})
	}

	return &entity.UpdateProviderToolResponse{
		Tool: modelmapper.FromProviderToolModelToProviderToolEntity(*toolRes),
	}, nil
}

func (c *controller) DeleteProviderTool(ctx context.Context, req entity.DeleteProviderToolRequest) error {
	// Init transaction
	ctx = c.storage.ContextWithTx(ctx)

	serverIDs, err := c.storage.ListServerIDsByToolID(ctx, req.ToolID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by tool id")
	}

	err = c.storage.DeleteProviderTool(ctx, req.ToolID)
	if err != nil {
		_ = c.storage.TxRollback(ctx)
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete provider tool",
			Data: map[string]any{
				"toolID":     req.ToolID,
				"providerID": req.ProviderID,
				"reason":     err.Error(),
			},
		}
	}

	// Updates version!
	err = c.storage.UpdateProvider(ctx, req.ProviderID, nil)
	if err != nil {
		_ = c.storage.TxRollback(ctx)
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update provider version due to tool deletion",
			Data: map[string]any{
				"reason":     err.Error(),
				"toolID":     req.ToolID,
				"providerID": req.ProviderID,
			},
		}
	}

	err = c.storage.TxCommit(ctx)
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "db transaction failed to update provider tool deletion",
			Data: map[string]any{
				"reason":     err.Error(),
				"toolID":     req.ToolID,
				"providerID": req.ProviderID,
			},
		}
	}

	freshCtx := context.Background()
	c.cache.Evict(freshCtx, entity.ObjectTypeProviderTool, req.ToolID)
	for _, id := range serverIDs {
		c.cache.Evict(freshCtx, entity.ObjectTypeServer, id)
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeProviderTool,
			EventType:       entity.ObjectEventTypeDelete,
			ResoureID:       req.ToolID,
			ResourceOwnerID: id,
		})
	}

	return nil
}

func (c *controller) validateCreateProviderToolRequest(req entity.CreateProviderToolRequest) error {
	e := req.Tool
	if e.Method == entity.MethodTypeInvalid {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("invalid method type: %s", e.Method),
			Data: map[string]any{
				"method": e.Method,
			},
		}
	}

	if len(e.Path) < _validationAttrProviderToolPathMinLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("path must be at least %d characters", _validationAttrProviderToolPathMinLength),
			Data: map[string]any{
				"path": e.Path,
			},
		}
	}
	if len(e.Path) > _validationAttrProviderToolPathMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("path exceeds maximum length of %d", _validationAttrProviderToolPathMaxLength),
			Data: map[string]any{
				"path": e.Path,
			},
		}
	}
	if len(e.Name) > 0 && !_regexValidationProviderToolName.MatchString(e.Name) {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "name must be 1-20 chars and start with smallcaps",
			Data: map[string]any{
				"name": e.Name,
			},
		}
	}

	if len(e.Title) > _validationAttrProviderToolTitleMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("title exceeds maximum length of %d", _validationAttrProviderToolTitleMaxLength),
			Data: map[string]any{
				"title": e.Title,
			},
		}
	}

	if len(e.Description) > _validationAttrProviderToolDescMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("description exceeds maximum length of %d", _validationAttrProviderToolDescMaxLength),
			Data: map[string]any{
				"description": e.Description,
			},
		}
	}

	compiler := jsonschema.NewCompiler()
	if len(e.PathArgsJSONSchema) > 0 {
		_, err := compiler.Compile(e.PathArgsJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid path arguments JSON schema for provider tool",
				Data: map[string]any{
					"reason":             err.Error(),
					"pathArgsJSONSchema": string(e.PathArgsJSONSchema),
				},
			}
		}
	}

	if len(e.QueryArgsJSONSchema) > 0 {
		_, err := compiler.Compile(e.QueryArgsJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid query arguments JSON schema for provider tool",
				Data: map[string]any{
					"reason":              err.Error(),
					"queryArgsJSONSchema": string(e.QueryArgsJSONSchema),
				},
			}
		}
	}

	if len(e.ReqBodyJSONSchema) > 0 {
		_, err := compiler.Compile(e.ReqBodyJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid request body JSON schema for provider tool",
				Data: map[string]any{
					"reason":            err.Error(),
					"reqBodyJSONSchema": string(e.ReqBodyJSONSchema),
				},
			}
		}
	}

	if len(e.ResBodyJSONSchema) > 0 {
		_, err := compiler.Compile(e.ResBodyJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid response body JSON schema for provider tool",
				Data: map[string]any{
					"reason":            err.Error(),
					"resBodyJSONSchema": string(e.ResBodyJSONSchema),
				},
			}
		}
	}

	return nil
}

func (c *controller) validateUpdateProviderToolRequest(req entity.UpdateProviderToolRequest) error {
	e := req.Tool
	var anyChanges bool
	if len(e.Description) > 0 {
		anyChanges = true
		if len(e.Description) > _validationAttrProviderToolDescMaxLength {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: fmt.Sprintf("description exceeds maximum length of %d", _validationAttrProviderToolDescMaxLength),
				Data: map[string]any{
					"description": e.Description,
				},
			}
		}
	}

	if len(e.Name) > 0 && !_regexValidationProviderToolName.MatchString(e.Name) {
		anyChanges = true
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "name must be 1-20 chars and start with smallcaps",
			Data: map[string]any{
				"name": e.Name,
			},
		}
	}

	if len(e.Title) > 0 {
		anyChanges = true
		if len(e.Title) > _validationAttrProviderToolTitleMaxLength {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: fmt.Sprintf("title exceeds maximum length of %d", _validationAttrProviderToolTitleMaxLength),
				Data: map[string]any{
					"title": e.Title,
				},
			}
		}
	}

	compiler := jsonschema.NewCompiler()
	if len(e.PathArgsJSONSchema) > 0 {
		_, err := compiler.Compile(e.PathArgsJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid path arguments JSON schema for provider tool",
				Data: map[string]any{
					"reason":             err.Error(),
					"pathArgsJSONSchema": string(e.PathArgsJSONSchema),
				},
			}
		}
		anyChanges = true
	}

	if len(e.QueryArgsJSONSchema) > 0 {
		_, err := compiler.Compile(e.QueryArgsJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid query arguments JSON schema for provider tool",
				Data: map[string]any{
					"reason":              err.Error(),
					"queryArgsJSONSchema": string(e.QueryArgsJSONSchema),
				},
			}
		}
		anyChanges = true
	}

	if len(e.ReqBodyJSONSchema) > 0 {
		_, err := compiler.Compile(e.ReqBodyJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid request body JSON schema for provider tool",
				Data: map[string]any{
					"reason":            err.Error(),
					"reqBodyJSONSchema": string(e.ReqBodyJSONSchema),
				},
			}
		}
		anyChanges = true
	}

	if len(e.ResBodyJSONSchema) > 0 {
		_, err := compiler.Compile(e.ResBodyJSONSchema)
		if err != nil {
			return erre.Error{
				Code:    erre.ErrorCodeBadRequest,
				Message: "invalid response body JSON schema for provider tool",
				Data: map[string]any{
					"reason":            err.Error(),
					"resBodyJSONSchema": string(e.ResBodyJSONSchema),
				},
			}
		}
		anyChanges = true
	}

	if !anyChanges {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "no changes provided for provider tool update",
		}
	}
	return nil
}
