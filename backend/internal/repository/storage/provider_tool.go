package storage

import (
	"context"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type ProviderToolStorage interface {
	CreateProviderTool(ctx context.Context, pe model.ProviderTool) error
	GetProviderTool(ctx context.Context, id int64) (*model.ProviderTool, error)
	ListProviderTools(ctx context.Context, providerID int64, ids []int64) ([]model.ProviderTool, error)
	UpdateProviderTool(ctx context.Context, id int64, attrs map[model.ProviderToolAttribute]any) error
	DeleteProviderTool(ctx context.Context, id int64) error

	// Admistrative
	ListTools(ctx context.Context, ids []int64) ([]model.ProviderTool, error)
}

func (r *repository) CreateProviderTool(ctx context.Context, pe model.ProviderTool) error {
	err := r.db.Conn(ctx).Create(&pe).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetProviderTool(ctx context.Context, id int64) (*model.ProviderTool, error) {
	var tool model.ProviderTool
	err := r.db.Conn(ctx).Where("id = ?", id, id).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *repository) ListProviderTools(ctx context.Context, providerID int64, ids []int64) ([]model.ProviderTool, error) {
	var tools []model.ProviderTool
	db := r.db.Conn(ctx).Where("provider_id = ?", providerID)
	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}
	err := db.Find(&tools).Error
	if err != nil {
		return nil, err
	}
	return tools, nil
}

func (r *repository) UpdateProviderTool(ctx context.Context, id int64, attrs map[model.ProviderToolAttribute]any) error {
	attrsModified := make(map[string]any, len(attrs))
	for k, v := range attrs {
		attrsModified[k.String()] = v
	}
	attrsModified[model.ProviderToolAttributeUpdatedAt.String()] = time.Now().UTC()
	err := r.db.Conn(ctx).Model(&model.ProviderTool{}).Where("id = ?", id).Updates(attrsModified).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) DeleteProviderTool(ctx context.Context, id int64) error {
	var err error
	db := r.db.Conn(ctx)
	// Delete associations
	err = db.Where("tool_id = ?", id).Delete(&model.ServerTool{}).Error
	if err != nil {
		return err
	}

	err = db.Where("id = ?", id).Delete(&model.ProviderTool{}).Error
	if err != nil {
		return err
	}
	return nil
}

// Administrative methods

func (r *repository) ListTools(ctx context.Context, ids []int64) ([]model.ProviderTool, error) {
	var tools []model.ProviderTool
	db := r.db.Conn(ctx)
	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}
	err := db.Find(&tools).Error
	if err != nil {
		return nil, err
	}
	return tools, nil
}
