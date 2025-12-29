package storage

import (
	"context"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type ServerToolStorage interface {
	AddToolToServer(ctx context.Context, e model.ServerTool) error
	RemoveServerTool(ctx context.Context, e model.ServerTool) error
	ListServerTools(ctx context.Context, serverID int64) ([]model.ServerTool, error)
	DeleteAllServerTools(ctx context.Context, serverID int64) error
	ListServerIDsByToolID(ctx context.Context, toolID int64) ([]int64, error)
	ListServerIDsByProviderID(ctx context.Context, providerID int64) ([]int64, error)
}

func (r *repository) AddToolToServer(ctx context.Context, e model.ServerTool) error {
	err := r.db.Conn(ctx).Create(&e).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) RemoveServerTool(ctx context.Context, e model.ServerTool) error {
	err := r.db.Conn(ctx).
		Where("server_id = ?", e.ServerID).
		Where("tool_id = ?", e.ToolID).
		Delete(&model.ServerTool{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) DeleteAllServerTools(ctx context.Context, serverID int64) error {
	err := r.db.Conn(ctx).
		Where("server_id = ?", serverID).
		Delete(&model.ServerTool{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) ListServerIDsByToolID(ctx context.Context, toolID int64) ([]int64, error) {
	var serverIDs []int64
	err := r.db.Conn(ctx).
		Table("server_tools").
		Where("tool_id = ?", toolID).
		Pluck("server_id", &serverIDs).
		Error
	if err != nil {
		return nil, err
	}

	return serverIDs, nil
}

func (r *repository) ListServerIDsByProviderID(ctx context.Context, providerID int64) ([]int64, error) {
	var serverIDs []int64
	err := r.db.Conn(ctx).
		Table("server_tools").
		Where("provider_id = ?", providerID).
		Pluck("server_id", &serverIDs).
		Error
	if err != nil {
		return nil, err
	}

	return serverIDs, nil
}
