package storage

import (
	"context"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type ServerPromptStorage interface {
	AddPromptToServer(ctx context.Context, e model.ServerPrompt) error
	RemoveServerPrompt(ctx context.Context, e model.ServerPrompt) error
	ListServerPrompts(ctx context.Context, serverID int64) ([]model.ServerPrompt, error)
	DeleteAllServerPrompts(ctx context.Context, serverID int64) error
	ListServerIDsByPromptID(ctx context.Context, promptID int64) ([]int64, error)
}

// ServerPrompt methods
func (r *repository) AddPromptToServer(ctx context.Context, e model.ServerPrompt) error {
	return r.db.Conn(ctx).Create(&e).Error
}

func (r *repository) RemoveServerPrompt(ctx context.Context, e model.ServerPrompt) error {
	return r.db.Conn(ctx).
		Where("server_id = ?", e.ServerID).
		Where("prompt_id = ?", e.PromptID).
		Delete(&model.ServerPrompt{}).Error
}

func (r *repository) ListServerPrompts(ctx context.Context, serverID int64) ([]model.ServerPrompt, error) {
	var prompts []model.ServerPrompt
	err := r.db.Conn(ctx).Where("server_id = ?", serverID).Find(&prompts).Error
	return prompts, err
}

func (r *repository) DeleteAllServerPrompts(ctx context.Context, serverID int64) error {
	return r.db.Conn(ctx).
		Where("server_id = ?", serverID).
		Delete(&model.ServerPrompt{}).Error
}

func (r *repository) ListServerIDsByPromptID(ctx context.Context, promptID int64) ([]int64, error) {
	var serverIDs []int64
	err := r.db.Conn(ctx).
		Table("server_prompts").
		Where("prompt_id = ?", promptID).
		Pluck("server_id", &serverIDs).
		Error
	if err != nil {
		return nil, err
	}

	return serverIDs, nil
}
