package transactional

import (
	"context"

	"github.com/go-pg/pg/v10"
)

// The contextKey type is unexported to prevent collisions with context keys defined in
// other packages
type contextKey string

// ValueFromContext retrives Value from context
func ValueFromContext(ctx context.Context, keyName string) interface{} {
	return ctx.Value(contextKey(keyName))
}

// ContextWithValue sets Value to context request
func ContextWithValue(ctx context.Context, keyName string, value interface{}) context.Context {
	return context.WithValue(ctx, contextKey(keyName), value)
}

// DbFromContext retrives Db from context
func DbFromContext(ctx context.Context) *pg.DB {
	v := ValueFromContext(ctx, "db")
	if v == nil {
		return nil
	}
	return v.(*pg.DB)
}

// ContextWithDb sets Db to context request
func ContextWithDb(ctx context.Context, db *pg.DB) context.Context {
	return context.WithValue(ctx, contextKey("db"), db)
}

// TxFromContext retrives Tx from context
func TxFromContext(ctx context.Context) *pg.Tx {
	v := ValueFromContext(ctx, "tx")
	if v == nil {
		return nil
	}
	return v.(*pg.Tx)
}

// ContextWithTx sets Tx to context request
func ContextWithTx(ctx context.Context, tx *pg.Tx) context.Context {
	return context.WithValue(ctx, contextKey("tx"), tx)
}
