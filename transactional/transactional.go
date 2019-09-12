package transactional

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v9"
)

type ExecFunc func(ctx context.Context, tx *pg.Tx) (context.Context, error)

// Propagation type
type Propagation string

const (
	// Support a current transaction, create a new one if none exists
	Required Propagation = "Required"

	// Support a current transaction, return an exception if none exists
	Mandatory Propagation = "Mandatory"

	//Create a new transaction, and suspend the current transaction if one exists
	RequiredNew Propagation = "RequiredNew"
)

// Execute executes ExecFunc in transaction
func Execute(ctx context.Context, execFunc ExecFunc) (context.Context, error) {
	return execute(ctx, Required, execFunc)
}

// ExecuteWithPropagation executes ExecFunc in transaction with specific propagation
func ExecuteWithPropagation(ctx context.Context, propagation Propagation, execFunc ExecFunc) (context.Context, error) {
	return execute(ctx, propagation, execFunc)
}

func execute(ctx context.Context, propagation Propagation, execFunc ExecFunc) (context.Context, error) {
	var err error
	var tx *pg.Tx
	var localtx *pg.Tx
	if propagation == RequiredNew {
		db := DbFromContext(ctx)
		if db == nil {
			return ctx, errors.New("No pg.DB found in context")
		}
		localtx, err = db.Begin()
		if err != nil {
			return ctx, err
		}
		tx = localtx
	} else {
		tx = TxFromContext(ctx)
		if tx == nil {
			if propagation == Mandatory {
				return ctx, errors.New("No pg.Tx found in context with Mandatory propagation")
			}
			db := DbFromContext(ctx)
			if db == nil {
				return ctx, errors.New("No pg.DB found in context")
			}
			localtx, err = db.Begin()
			if err != nil {
				return ctx, err
			}
			tx = localtx
		}
	}
	ctx = ContextWithTx(ctx, tx)
	ctx, err = execFunc(ctx, tx)
	if localtx != nil {
		if err == nil {
			err = localtx.Commit()
		} else {
			localtx.Rollback()
		}
		ctx = ContextWithTx(ctx, nil)
	}
	return ctx, err
}
