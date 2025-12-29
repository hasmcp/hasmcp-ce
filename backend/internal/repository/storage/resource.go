package storage

import (
	"context"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type ResourceStorage interface {
	CreateResource(ctx context.Context, r model.Resource) error
	GetResource(ctx context.Context, id int64) (*model.Resource, error)
	ListResources(ctx context.Context, resourceIDs []int64) ([]model.Resource, error)
	DeleteResource(ctx context.Context, id int64) error
	UpdateResource(ctx context.Context, id int64, attrs map[model.ResourceAttribute]any) error
}

// CreateResource creates a new resource in storage.
func (r *repository) CreateResource(ctx context.Context, res model.Resource) error {
	res.CreatedAt = time.Now().UTC()
	res.UpdatedAt = res.CreatedAt
	err := r.db.Conn(ctx).Create(&res).Error
	if err != nil {
		return err
	}
	return nil
}

// ListResources lists all resources from storage, optionally filtered by IDs.
func (r *repository) ListResources(ctx context.Context, resourceIDs []int64) ([]model.Resource, error) {
	var resources []model.Resource
	db := r.db.Conn(ctx)
	if len(resourceIDs) > 0 {
		db = db.Where("id IN ?", resourceIDs)
	}
	err := db.Find(&resources).Error
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// GetResource finds a resource by its ID.
func (r *repository) GetResource(ctx context.Context, id int64) (*model.Resource, error) {
	var resource model.Resource
	err := r.db.Conn(ctx).Where("id = ?", id).First(&resource).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// DeleteResource deletes a resource by its ID.
func (r *repository) DeleteResource(ctx context.Context, id int64) error {
	var err error
	db := r.db.Conn(ctx)
	// Delete associations

	err = db.Where("resource_id = ?", id).Delete(&model.ServerResource{}).Error
	if err != nil {
		return err
	}

	err = db.Where("id = ?", id).Delete(&model.Resource{}).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateResource updates a resource by its ID.
func (r *repository) UpdateResource(ctx context.Context, id int64, attrs map[model.ResourceAttribute]any) error {
	attrsModified := make(map[string]any, len(attrs))
	for k, v := range attrs {
		attrsModified[k.String()] = v
	}
	attrsModified[model.ResourceAttributeUpdatedAt.String()] = time.Now().UTC()
	err := r.db.Conn(ctx).Model(&model.Resource{}).Where("id = ?", id).Updates(attrsModified).Error
	if err != nil {
		return err
	}
	return nil
}
