package storage

import (
	"context"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/base"
)

type (
	Params struct {
		DB base.Repository
	}

	txRepository interface {
		ContextWithTx(ctx context.Context) context.Context
		TxCommit(ctx context.Context) error
		TxRollback(ctx context.Context) error
	}

	Repository interface {
		txRepository

		VariableStorage

		ProviderStorage
		ProviderToolStorage

		PromptStorage

		ResourceStorage

		ServerStorage
		ServerToolStorage
		ServerPromptStorage
		ServerResourceStorage
	}

	repository struct {
		db base.Repository
	}
)

// New initializes the storage repository.
func New(p Params) (Repository, error) {
	ctx := context.Background()
	if err := p.DB.Conn(ctx).AutoMigrate(&model.Variable{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.Provider{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.ProviderTool{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.ProviderOauth2Config{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.Resource{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.Prompt{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.Server{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.ServerTool{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.ServerPrompt{}); err != nil {
		return nil, err
	}

	if err := p.DB.Conn(ctx).AutoMigrate(&model.ServerResource{}); err != nil {
		return nil, err
	}

	return &repository{
		db: p.DB,
	}, nil
}

func (r *repository) ContextWithTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, base.CtxKeyDBTx, r.db.TxBegin(ctx))
}

func (r *repository) TxCommit(ctx context.Context) error {
	return r.db.TxCommit(ctx)
}

func (r *repository) TxRollback(ctx context.Context) error {
	return r.db.TxRollback(ctx)
}
