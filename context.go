package pgrest

import (
	"context"

	"github.com/go-pg/pg/v9"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

const (
	// dbKey is the context key for DB. Its value of zero is
	// arbitrary.  If this package defined other context keys, they would have
	// different integer values.
	dbKey = iota

	// txKey is the context key for Tx.
	txKey
)

// DbFromContext retrives Db from context
func DbFromContext(ctx context.Context) *pg.DB {
	v := ctx.Value(dbKey)
	if v == nil {
		return nil
	}
	return v.(*pg.DB)
}

// SetDbToContext sets Db to context request
func SetDbToContext(ctx context.Context, db *pg.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

// TxFromContext retrives Tx from context
func TxFromContext(ctx context.Context) *pg.Tx {
	v := ctx.Value(txKey)
	if v == nil {
		return nil
	}
	return v.(*pg.Tx)
}

// SetTxToContext sets Tx to context request
func SetTxToContext(ctx context.Context, tx *pg.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}
