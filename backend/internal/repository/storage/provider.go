package storage

import (
	"context"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"gorm.io/gorm"
)

type ProviderStorage interface {
	CreateProvider(ctx context.Context, p model.Provider) error
	GetProvider(ctx context.Context, id int64) (*model.Provider, error)
	ListProviders(ctx context.Context, providerIDs []int64) ([]model.Provider, error)
	UpdateProvider(ctx context.Context, id int64, attrs map[model.ProviderAttribute]any) error
	DeleteProvider(ctx context.Context, id int64) error
}

func (r *repository) CreateProvider(ctx context.Context, p model.Provider) error {
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	err := r.db.Conn(ctx).Create(&p).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetProvider(ctx context.Context, id int64) (*model.Provider, error) {
	var provider model.Provider
	err := r.db.Conn(ctx).Preload("Tools").Preload("Oauth2Config").Where("id = ?", id).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *repository) ListProviders(ctx context.Context, providerIDs []int64) ([]model.Provider, error) {
	var providers []model.Provider
	db := r.db.Conn(ctx)
	if len(providerIDs) > 0 {
		db = db.Where("id IN ?", providerIDs)
	}
	err := db.Find(&providers).Error
	if err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *repository) UpdateProvider(ctx context.Context, id int64, attrs map[model.ProviderAttribute]any) error {
	db := r.db.Conn(ctx)
	if v, ok := attrs[model.ProviderAttributeOauth2Config]; ok {
		err := db.Model(&model.ProviderOauth2Config{}).Where("id = ?", id).Save(v).Error
		if err != nil {
			return err
		}
		delete(attrs, model.ProviderAttributeOauth2Config)
	}

	attrsModified := make(map[string]any, len(attrs))
	for k, v := range attrs {
		attrsModified[k.String()] = v
	}
	attrsModified[model.ProviderAttributeUpdatedAt.String()] = time.Now().UTC()
	attrsModified[model.ProviderAttributeVersion.String()] = gorm.Expr("version + ?", 1)
	err := r.db.Conn(ctx).Model(&model.Provider{}).Where("id = ?", id).Updates(attrsModified).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) DeleteProvider(ctx context.Context, id int64) error {
	var err error
	db := r.db.Conn(ctx)
	// Delete associations
	err = db.Where("provider_id = ?", id).Delete(&model.ProviderOauth2Config{}).Error
	if err != nil {
		return err
	}

	err = db.Where("provider_id = ?", id).Delete(&model.ProviderTool{}).Error
	if err != nil {
		return err
	}

	err = db.Where("provider_id = ?", id).Delete(&model.ServerTool{}).Error
	if err != nil {
		return err
	}

	err = db.Where("id = ?", id).Delete(&model.Provider{}).Error
	if err != nil {
		return err
	}

	return nil
}
