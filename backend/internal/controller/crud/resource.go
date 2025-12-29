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

type ResourceController interface {
	CreateResource(ctx context.Context, req entity.CreateResourceRequest) (*entity.CreateResourceResponse, error)
	GetResource(ctx context.Context, req entity.GetResourceRequest) (*entity.GetResourceResponse, error)
	ListResources(ctx context.Context, req entity.ListResourcesRequest) (*entity.ListResourcesResponse, error)
	UpdateResource(ctx context.Context, req entity.UpdateResourceRequest) error
	DeleteResource(ctx context.Context, req entity.DeleteResourceRequest) error
}

const (
	_validationAttrResourceNameMaxLength        = 128
	_validationAttrResourceDescriptionMaxLength = 1024
	_validationAttrResourceURIMaxLength         = 255
)

func (c *controller) CreateResource(ctx context.Context, req entity.CreateResourceRequest) (*entity.CreateResourceResponse, error) {
	if err := c.validateCreateResourceRequest(req); err != nil {
		return nil, err
	}

	r := req.Resource

	now := time.Now().UTC()
	res := model.Resource{
		ID:          c.idgen.Next(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Name:        r.Name,
		Description: r.Description,
		URI:         r.URI,
		MimeType:    r.MimeType,
		Size:        r.Size,
		Annotations: r.Annotations,
	}

	err := c.storage.CreateResource(ctx, res)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create server resource",
			Data: map[string]any{
				"reason": err.Error(),
				"name":   r.Name,
				"uri":    r.URI,
			},
		}
	}

	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType: entity.ObjectTypeResource,
		EventType:  entity.ObjectEventTypeCreate,
		ResoureID:  res.ID,
	})

	return &entity.CreateResourceResponse{Resource: modelmapper.FromResourceModelToReourceEntity(res)}, nil
}

func (c *controller) GetResource(ctx context.Context, req entity.GetResourceRequest) (*entity.GetResourceResponse, error) {
	res, err := c.storage.GetResource(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeNotFound,
				Message: "server resource not found",
				Data:    map[string]any{"id": req.ID},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get server resource",
			Data:    map[string]any{"id": req.ID, "reason": err.Error()},
		}
	}

	return &entity.GetResourceResponse{
		Resource: modelmapper.FromResourceModelToReourceEntity(*res),
	}, nil
}

func (c *controller) ListResources(ctx context.Context, req entity.ListResourcesRequest) (*entity.ListResourcesResponse, error) {
	resources, err := c.storage.ListResources(ctx, nil)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to list server resources",
			Data:    map[string]any{"reason": err.Error()},
		}
	}

	return &entity.ListResourcesResponse{
		Resources: modelmapper.FromResourceModelsToResourceEntities(resources),
	}, nil
}

func (c *controller) UpdateResource(ctx context.Context, req entity.UpdateResourceRequest) error {
	if err := c.validateUpdateResourceRequest(req); err != nil {
		return err
	}
	r := req.Resource

	attrs := make(map[model.ResourceAttribute]any)
	if r.Name != "" {
		attrs[model.ResourceAttributeName] = r.Name
	}
	if r.Description != "" {
		attrs[model.ResourceAttributeDescription] = r.Description
	}
	if r.URI != "" {
		attrs[model.ResourceAttributeURI] = r.URI
	}
	if r.MimeType != "" {
		attrs[model.ResourceAttributeMimeType] = r.MimeType
	}
	if r.Size != 0 {
		attrs[model.ResourceAttributeSize] = r.Size
	}
	if r.Annotations != nil {
		attrs[model.ResourceAttributeAnnotations] = r.Annotations
	}

	if len(attrs) == 0 {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "no update fields provided",
			Data:    map[string]any{"id": r.ID},
		}
	}

	err := c.storage.UpdateResource(ctx, r.ID, attrs)
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update server resource",
			Data:    map[string]any{"id": r.ID, "reason": err.Error()},
		}
	}

	serverIDs, err := c.storage.ListServerIDsByResourceID(ctx, r.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by resource")
	}

	c.cache.Evict(ctx, entity.ObjectTypeResource, req.Resource.ID)
	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeResource,
			EventType:       entity.ObjectEventTypeUpdate,
			ResoureID:       req.Resource.ID,
			ResourceOwnerID: id,
		})
	}

	return nil
}

func (c *controller) DeleteResource(ctx context.Context, req entity.DeleteResourceRequest) error {
	defer c.cache.Evict(ctx, entity.ObjectTypeResource, req.ID)

	serverIDs, err := c.storage.ListServerIDsByResourceID(ctx, req.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by resource")
	}

	err = c.storage.DeleteResource(ctx, req.ID)
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete server resource",
			Data:    map[string]any{"id": req.ID, "reason": err.Error()},
		}
	}

	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeResource,
			EventType:       entity.ObjectEventTypeDelete,
			ResoureID:       req.ID,
			ResourceOwnerID: id,
		})
	}

	return nil
}

func (c *controller) validateCreateResourceRequest(req entity.CreateResourceRequest) error {
	r := req.Resource
	if r.Name == "" {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "resource name is required"}
	}
	if len(r.Name) > _validationAttrResourceNameMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("resource name exceeds maximum length of %d", _validationAttrResourceNameMaxLength)}
	}
	if r.URI == "" {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "resource URI is required"}
	}
	if len(r.URI) > _validationAttrResourceURIMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("resource URI exceeds maximum length of %d", _validationAttrResourceURIMaxLength)}
	}
	if len(r.Description) > _validationAttrResourceDescriptionMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("resource description exceeds maximum length of %d", _validationAttrResourceDescriptionMaxLength)}
	}
	return nil
}

func (c *controller) validateUpdateResourceRequest(req entity.UpdateResourceRequest) error {
	r := req.Resource
	if r.ID == 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "resource ID is required"}
	}
	if len(r.Name) > _validationAttrResourceNameMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("resource name exceeds maximum length of %d", _validationAttrResourceNameMaxLength)}
	}
	if len(r.URI) > _validationAttrResourceURIMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("resource URI exceeds maximum length of %d", _validationAttrResourceURIMaxLength)}
	}
	if len(r.Description) > _validationAttrResourceDescriptionMaxLength {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: fmt.Sprintf("resource description exceeds maximum length of %d", _validationAttrResourceDescriptionMaxLength)}
	}
	return nil
}
