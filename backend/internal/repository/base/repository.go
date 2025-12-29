package base

import (
	"context"

	"gorm.io/gorm"
)

type (
	Repository interface {
		Conn(ctx context.Context) *gorm.DB
		Close(ctx context.Context)

		TxBegin(ctx context.Context) *gorm.DB
		TxCommit(ctx context.Context) error
		TxRollback(ctx context.Context) error
	}

	ctxKey uint8
)

const (
	// CtxKeyDBTx is the context key for the database transaction
	CtxKeyDBTx ctxKey = iota
)

// ErrRecordNotFound is a sentinel error that indicates that a record was not found
var ErrRecordNotFound = gorm.ErrRecordNotFound
