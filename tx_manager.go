package txManager

import (
	"context"

	"github.com/pkg/errors"
)

type Manager struct{}

func NewTxManager() *Manager {
	return &Manager{}
}

type Tx struct {
	transactors []Transactor
}

func (m *Manager) NewTransaction(transactors ...Transactor) Tx {
	return Tx{transactors: transactors}
}

func (tx Tx) Execute(ctx context.Context, f func(ctx context.Context) error) (err error) {
	internalTransactions := make([]Transaction, 0, len(tx.transactors))

	for _, tr := range tx.transactors {
		ctxTx, internalTx, err := tr.BeginTxWithContext(ctx)
		if err != nil {
			return errors.WithStack(errors.Wrap(err, "failed to begin internal tx"))
		}

		internalTransactions = append(internalTransactions, internalTx)
		ctx = ctxTx
	}

	defer func() {
		defer func() {
			tx.transactors = nil
		}()

		if r := recover(); r != nil {
			err = errors.Wrapf(err, "panic recovered: %v", r)
		}

		if err != nil {
			for _, tx := range internalTransactions {
				if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
					err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
				}
			}
			return
		}

		for _, tx := range internalTransactions {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = errors.Wrapf(err, "commit error: %v", commitErr)
			}
		}
	}()

	if err := f(ctx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
