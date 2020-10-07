package criusctx

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	transaction = "Transaction"
)

func WithTransaction(ctx context.Context, txn *sqlx.Tx) context.Context {
	return context.WithValue(ctx, transaction, txn)
}

func GetTransaction(ctx context.Context) (boil.ContextExecutor, bool) {
	value := ctx.Value(transaction)
	if value == nil {
		return nil, false
	}
	exec, ok := value.(boil.ContextExecutor)
	return exec, ok
}
