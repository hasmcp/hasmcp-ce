package storage

import (
	"context"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type PromptStorage interface {
	CreatePrompt(ctx context.Context, p model.Prompt) error
	GetPrompt(ctx context.Context, id int64) (*model.Prompt, error)
	ListPrompts(ctx context.Context, promptIDs []int64) ([]model.Prompt, error)
	DeletePrompt(ctx context.Context, id int64) error
	UpdatePrompt(ctx context.Context, id int64, attrs map[model.PromptAttribute]any) error
}

// CreatePrompt creates a new prompt in storage.
func (r *repository) CreatePrompt(ctx context.Context, p model.Prompt) error {
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	err := r.db.Conn(ctx).Create(&p).Error
	if err != nil {
		return err
	}
	return nil
}

// ListPrompts lists all prompts from storage, optionally filtered by IDs.
func (r *repository) ListPrompts(ctx context.Context, promptIDs []int64) ([]model.Prompt, error) {
	var prompts []model.Prompt
	db := r.db.Conn(ctx)
	if len(promptIDs) > 0 {
		db = db.Where("id IN ?", promptIDs)
	}
	err := db.Find(&prompts).Error
	if err != nil {
		return nil, err
	}
	return prompts, nil
}

// GetPrompt finds a prompt by its ID.
func (r *repository) GetPrompt(ctx context.Context, id int64) (*model.Prompt, error) {
	var prompt model.Prompt
	err := r.db.Conn(ctx).Where("id = ?", id).First(&prompt).Error
	if err != nil {
		return nil, err
	}
	return &prompt, nil
}

// DeletePrompt deletes a prompt by its ID.
func (r *repository) DeletePrompt(ctx context.Context, id int64) error {
	var err error
	db := r.db.Conn(ctx)
	// Delete associations

	err = db.Where("prompt_id = ?", id).Delete(&model.ServerPrompt{}).Error
	if err != nil {
		return err
	}

	err = r.db.Conn(ctx).Where("id = ?", id).Delete(&model.Prompt{}).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdatePrompt updates a prompt by its ID.
func (r *repository) UpdatePrompt(ctx context.Context, id int64, attrs map[model.PromptAttribute]any) error {
	attrsModified := make(map[string]any, len(attrs))
	for k, v := range attrs {
		attrsModified[k.String()] = v
	}
	attrsModified[model.PromptAttributeUpdatedAt.String()] = time.Now().UTC()
	err := r.db.Conn(ctx).Model(&model.Prompt{}).Where("id = ?", id).Updates(attrsModified).Error
	if err != nil {
		return err
	}
	return nil
}
