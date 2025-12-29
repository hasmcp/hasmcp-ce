package storage

import (
	"context"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type ServerResourceStorage interface {
	AddResourceToServer(ctx context.Context, e model.ServerResource) error
	RemoveServerResource(ctx context.Context, e model.ServerResource) error
	ListServerResources(ctx context.Context, serverID int64) ([]model.ServerResource, error)
	DeleteAllServerResources(ctx context.Context, serverID int64) error
	ListServerIDsByResourceID(ctx context.Context, resourceID int64) ([]int64, error)
}

// ServerResource methods
func (r *repository) AddResourceToServer(ctx context.Context, e model.ServerResource) error {
	return r.db.Conn(ctx).Create(&e).Error
}

func (r *repository) RemoveServerResource(ctx context.Context, e model.ServerResource) error {
	return r.db.Conn(ctx).
		Where("server_id = ?", e.ServerID).
		Where("resource_id = ?", e.ResourceID).
		Delete(&model.ServerResource{}).Error
}

func (r *repository) ListServerResources(ctx context.Context, serverID int64) ([]model.ServerResource, error) {
	var resources []model.ServerResource
	err := r.db.Conn(ctx).Where("server_id = ?", serverID).Find(&resources).Error
	return resources, err
}

func (r *repository) DeleteAllServerResources(ctx context.Context, serverID int64) error {
	return r.db.Conn(ctx).
		Where("server_id = ?", serverID).
		Delete(&model.ServerResource{}).Error
}

func (r *repository) ListServerIDsByResourceID(ctx context.Context, resourceID int64) ([]int64, error) {
	var serverIDs []int64
	err := r.db.Conn(ctx).
		Table("server_resources").
		Where("resource_id = ?", resourceID).
		Pluck("server_id", &serverIDs).
		Error
	if err != nil {
		return nil, err
	}

	return serverIDs, nil
}
