package storage

import (
	"context"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"gorm.io/gorm"
)

type ServerStorage interface {
	CreateServer(ctx context.Context, d model.Server) error
	GetServer(ctx context.Context, id int64) (*model.Server, error)
	ListServers(ctx context.Context) ([]model.Server, error)
	UpdateServer(ctx context.Context, id int64, version int32, attrs map[model.ServerAttribute]any) error
	DeleteServer(ctx context.Context, id int64) error
}

func (r *repository) CreateServer(ctx context.Context, d model.Server) error {
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = d.CreatedAt
	err := r.db.Conn(ctx).Create(&d).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetServer(ctx context.Context, id int64) (*model.Server, error) {
	var server model.Server
	err := r.db.Conn(ctx).Where("id = ?", id).Preload("Tools").Preload("Prompts").Preload("Resources").First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *repository) ListServerTools(ctx context.Context, serverID int64) ([]model.ServerTool, error) {
	var tools []model.ServerTool
	err := r.db.Conn(ctx).Where("server_id = ?", serverID).Find(&tools).Error
	if err != nil {
		return nil, err
	}

	return tools, nil
}

func (r *repository) ListServers(ctx context.Context) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Conn(ctx).Preload("Tools").Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *repository) UpdateServer(ctx context.Context, id int64, version int32, attrs map[model.ServerAttribute]any) error {
	attrsModified := make(map[string]any, len(attrs))
	for k, v := range attrs {
		attrsModified[k.String()] = v
	}
	attrsModified[model.ServerAttributeUpdatedAt.String()] = time.Now().UTC()
	attrsModified[model.ServerAttributeVersion.String()] = gorm.Expr("version + ?", 1)
	delete(attrsModified, model.ServerAttributeTools.String())
	delete(attrsModified, model.ServerAttributePrompts.String())
	delete(attrsModified, model.ServerAttributeResources.String())

	db := r.db.Conn(ctx)

	err := db.
		Model(&model.Server{}).
		Where("id = ?", id).Where("version = ?", version).
		Updates(attrsModified).Error
	if err != nil {
		return err
	}
	err = db.Model(&model.ServerTool{}).Where("server_id = ?", id).Delete(&model.ServerTool{}).Error
	if err != nil {
		return err
	}

	err = db.Model(&model.ServerPrompt{}).Where("server_id = ?", id).Delete(&model.ServerPrompt{}).Error
	if err != nil {
		return err
	}

	err = db.Model(&model.ServerResource{}).Where("server_id = ?", id).Delete(&model.ServerResource{}).Error
	if err != nil {
		return err
	}

	tools := attrs[model.ServerAttributeTools]
	if tools != nil && len(tools.([]model.ServerTool)) > 0 {
		err = db.Model(&model.ServerTool{}).Save(tools).Error
		if err != nil {
			return err
		}
	}

	prompts := attrs[model.ServerAttributePrompts]
	if prompts != nil && len(prompts.([]model.ServerPrompt)) > 0 {
		err = db.Model(&model.ServerPrompt{}).Save(prompts).Error
		if err != nil {
			return err
		}
	}

	resources := attrs[model.ServerAttributeResources]
	if resources != nil && len(resources.([]model.ServerResource)) > 0 {
		err = db.Model(&model.ServerResource{}).Save(resources).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) DeleteServer(ctx context.Context, id int64) error {
	var err error
	db := r.db.Conn(ctx)
	// Delete associations
	err = db.Where("server_id = ?", id).Delete(&model.ServerTool{}).Error
	if err != nil {
		return err
	}
	err = db.Where("server_id = ?", id).Delete(&model.ServerPrompt{}).Error
	if err != nil {
		return err
	}
	err = db.Where("server_id = ?", id).Delete(&model.ServerResource{}).Error
	if err != nil {
		return err
	}

	// Delete mcp server
	err = db.Where("id = ?", id).Delete(&model.Server{}).Error
	if err != nil {
		return err
	}
	return nil
}
