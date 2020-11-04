package transactional

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
)

// ExecFunc definition
type ExecFunc func(ctx context.Context, tx *pg.Tx) error

// Propagation type
type Propagation string

const (
	// Current supports a current transaction, creates a new one if none exists.
	Current Propagation = "Current"

	// Mandatory needs a current transaction, return an exception if none exists
	Mandatory Propagation = "Mandatory"

	// Savepoint supports a current transaction, creates a new one if none exists, creates savepoint and never return propagation error
	Savepoint Propagation = "Savepoint"
)

// PropagationError struct
type propagationError struct {
	Cause       error
	Propagation Propagation
}

// newPropagationError constructs PropagationError
func newPropagationError(cause error, propagation Propagation) *propagationError {
	return &propagationError{Cause: cause, Propagation: propagation}
}

// Error implements the error interface
func (e propagationError) Error() string {
	return e.Cause.Error()
}

// Execute executes ExecFunc in transaction
func Execute(ctx context.Context, execFunc ExecFunc) error {
	return execute(ctx, Current, execFunc)
}

// ExecuteWithPropagation executes ExecFunc in transaction with specific propagation
func ExecuteWithPropagation(ctx context.Context, propagation Propagation, execFunc ExecFunc) error {
	return execute(ctx, propagation, execFunc)
}

func execute(ctx context.Context, propagation Propagation, execFunc ExecFunc) error {
	var err error
	var localtx *pg.Tx
	var savepoint string
	defer func() {
		if localtx != nil {
			if err == nil || propagation == Savepoint {
				localtx.Commit()
			} else {
				localtx.Rollback()
			}
		}
	}()
	db := DbFromContext(ctx)
	tx := TxFromContext(ctx)
	if tx == nil {
		if propagation == Mandatory {
			return newPropagationError(errors.New("No pg.Tx found in context with Mandatory propagation"), propagation)
		}
		if db == nil {
			return newPropagationError(errors.New("No pg.DB found in context"), propagation)
		}
		tx, err = db.Begin()
		if err != nil {
			return newPropagationError(err, propagation)
		}
		localtx = tx
	}
	if propagation == Savepoint {
		savepoint = "sp" + strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(rand.Int63(), 16)
		_, err = tx.Exec("SAVEPOINT " + savepoint)
		if err != nil {
			// Never return propagation error for Savepoint
			return nil
		}
	}
	err = execFunc(ContextWithTx(ctx, tx), tx)
	if savepoint != "" {
		if err == nil {
			tx.Exec("RELEASE SAVEPOINT " + savepoint)
		} else {
			tx.Exec("ROLLBACK TO SAVEPOINT " + savepoint)
		}
		// Never return propagation error for Savepoint
		return nil
	}
	if err != nil {
		return newPropagationError(err, propagation)
	}
	return nil
}
