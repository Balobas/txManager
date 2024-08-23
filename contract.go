package transaction

import (
	"context"
)

type Transactor interface {
	BeginTxWithContext(ctx context.Context) (context.Context, Transaction, error)
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
