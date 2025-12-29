package crud

import (
	"context"
	"errors"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"gorm.io/gorm"
)

type ServerResourceController interface {
	CreateServerResource(ctx context.Context, req entity.CreateServerResourceRequest) (*entity.CreateServerResourceResponse, error)
	DeleteServerResource(ctx context.Context, req entity.DeleteServerResourceRequest) error
	ListServerResources(ctx context.Context, req entity.ListServerResourcesRequest) (*entity.ListServerResourcesResponse, error)
}

func (c *controller) CreateServerResource(
	ctx context.Context, req entity.CreateServerResourceRequest) (*entity.CreateServerResourceResponse, error) {
	if err := c.validateCreateServerResourceRequest(req); err != nil {
		return nil, err
	}

	r := req.Resource

	// Check if server exists
	if _, err := c.GetServer(ctx, entity.GetServerRequest{ID: r.ServerID}); err != nil {
		return nil, err
	}

	// Check if resource exists
	if _, err := c.storage.GetResource(ctx, r.ResourceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeNotFound,
				Message: "resource not found",
				Data:    map[string]any{"resourceID": r.ResourceID},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get resource",
			Data:    map[string]any{"resourceID": r.ResourceID, "reason": err.Error()},
		}
	}

	res := model.ServerResource{
		ServerID:   r.ServerID,
		ResourceID: r.ResourceID,
	}
	err := c.storage.AddResourceToServer(ctx, res)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, erre.Error{
				Code:    erre.ErrorCodeConflict,
				Message: "server resource association already exists",
				Data:    map[string]any{"serverID": r.ServerID, "resourceID": r.ResourceID},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create server resource association",
			Data:    map[string]any{"serverID": r.ServerID, "resourceID": r.ResourceID, "reason": err.Error()},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeServer, r.ServerID)
	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerResource,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       r.ResourceID,
		ResourceOwnerID: r.ServerID,
	})

	return &entity.CreateServerResourceResponse{}, nil
}

func (c *controller) DeleteServerResource(ctx context.Context, req entity.DeleteServerResourceRequest) error {
	if err := c.validateDeleteServerResourceRequest(req); err != nil {
		return err
	}
	err := c.storage.RemoveServerResource(ctx, model.ServerResource{
		ServerID:   req.ServerID,
		ResourceID: req.ResourceID,
	})
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete server resource association",
			Data:    map[string]any{"serverID": req.ServerID, "resourceID": req.ResourceID, "reason": err.Error()},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeServer, req.ServerID)
	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerResource,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       req.ResourceID,
		ResourceOwnerID: req.ServerID,
	})

	return nil
}

func (c *controller) ListServerResources(ctx context.Context, req entity.ListServerResourcesRequest) (*entity.ListServerResourcesResponse, error) {
	if req.ServerID <= 0 {
		return nil, erre.Error{Code: erre.ErrorCodeBadRequest, Message: "invalid server ID"}
	}

	resources, err := c.storage.ListServerResources(ctx, req.ServerID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to list server resource associations",
			Data:    map[string]any{"serverID": req.ServerID, "reason": err.Error()},
		}
	}

	rs := make([]entity.ServerResource, len(resources))
	for i, r := range resources {
		rs[i] = entity.ServerResource{
			ServerID:   r.ServerID,
			ResourceID: r.ResourceID,
		}
	}

	return &entity.ListServerResourcesResponse{
		Resources: rs,
	}, nil
}

func (c *controller) validateCreateServerResourceRequest(req entity.CreateServerResourceRequest) error {
	r := req.Resource
	if r.ServerID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "server ID must be greater than 0"}
	}
	if r.ResourceID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "resource ID must be greater than 0"}
	}
	return nil
}

func (c *controller) validateDeleteServerResourceRequest(req entity.DeleteServerResourceRequest) error {
	if req.ServerID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "server ID must be greater than 0"}
	}
	if req.ResourceID <= 0 {
		return erre.Error{Code: erre.ErrorCodeBadRequest, Message: "resource ID must be greater than 0"}
	}
	return nil
}
