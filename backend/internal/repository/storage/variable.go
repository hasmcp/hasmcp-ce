package storage

import (
	"context"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

type VariableStorage interface {
	CreateVariable(ctx context.Context, v model.Variable) error
	SaveVariable(ctx context.Context, v model.Variable) error
	GetVariable(ctx context.Context, id int64) (*model.Variable, error)
	GetVariableByName(ctx context.Context, name string) (*model.Variable, error)
	ListVariables(ctx context.Context) ([]model.Variable, error)
	DeleteVariable(ctx context.Context, id int64) error
	UpdateVariable(ctx context.Context, id int64, attrs map[model.VariableAttribute]any) error
}

// CreateVariable creates a new variable in storage.
func (r *repository) CreateVariable(ctx context.Context, v model.Variable) error {
	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt
	err := r.db.Conn(ctx).Create(&v).Error
	if err != nil {
		return err
	}
	return nil
}

// SaveVariable saves a new variable in storage.
func (r *repository) SaveVariable(ctx context.Context, v model.Variable) error {
	err := r.db.Conn(ctx).Where("name = ?", v.Name).Delete(&model.Variable{}).Error
	if err != nil {
		return err
	}

	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt
	err = r.db.Conn(ctx).Create(&v).Error
	if err != nil {
		return err
	}
	return nil
}

// ListVariables lists all variables from storage.
func (r *repository) ListVariables(ctx context.Context) ([]model.Variable, error) {
	var variables []model.Variable
	err := r.db.Conn(ctx).Find(&variables).Error
	if err != nil {
		return nil, err
	}
	return variables, nil
}

// GetVariable finds a variable by its ID.
func (r *repository) GetVariable(ctx context.Context, id int64) (*model.Variable, error) {
	var variable model.Variable
	err := r.db.Conn(ctx).Where("id = ?", id).First(&variable).Error
	if err != nil {
		return nil, err
	}
	return &variable, nil
}

// GetVariable finds a variable by its ID.
func (r *repository) GetVariableByName(ctx context.Context, name string) (*model.Variable, error) {
	var variable model.Variable
	err := r.db.Conn(ctx).Where("name = ?", name).First(&variable).Error
	if err != nil {
		return nil, err
	}
	return &variable, nil
}

// DeleteVariable deletes a variable by its ID.
func (r *repository) DeleteVariable(ctx context.Context, id int64) error {
	err := r.db.Conn(ctx).Where("id = ?", id).Delete(&model.Variable{}).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateVariable updates a variable by its ID.
func (r *repository) UpdateVariable(ctx context.Context, id int64, attrs map[model.VariableAttribute]any) error {
	attrsModified := make(map[string]any, len(attrs))
	for k, v := range attrs {
		attrsModified[k.String()] = v
	}
	attrsModified[model.VariableAttributeUpdatedAt.String()] = time.Now().UTC()

	err := r.db.Conn(ctx).Model(&model.Variable{}).Where("id = ?", id).Updates(attrsModified).Error
	if err != nil {
		return err
	}
	return nil
}
