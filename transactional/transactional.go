package transactional

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-pg/pg/v9"
)

// ExecFunc definition
type ExecFunc func(ctx context.Context, tx *pg.Tx) error

// Propagation type
type Propagation string

const (
	// Current supports a current transaction, creates a new one if none exists
	Current Propagation = "Current"

	// Mandatory Current a current transaction, return an exception if none exists
	Mandatory Propagation = "Mandatory"

	// Savepoint supports a current transaction, creates a new one if none exists and creates savepoint
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
	db := DbFromContext(ctx)
	tx := TxFromContext(ctx)
	if tx == nil {
		if propagation == Mandatory {
			return newPropagationError(errors.New("No pg.Tx found in context with Mandatory propagation"), propagation)
		}
		if db == nil {
			return newPropagationError(errors.New("No pg.DB found in context"), propagation)
		}
		localtx, err = db.Begin()
		tx = localtx
		if err != nil {
			return newPropagationError(err, propagation)
		}
	}
	if propagation == Savepoint {
		savepoint = "sp" + strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(rand.Int63(), 16)
		_, err = tx.Exec("SAVEPOINT " + savepoint)
		if err != nil {
			return newPropagationError(err, propagation)
		}
	}
	err = execFunc(ContextWithTx(ctx, tx), tx)
	if savepoint != "" {
		if err == nil {
			tx.Exec("RELEASE SAVEPOINT " + savepoint)
		} else {
			tx.Exec("ROLLBACK TO SAVEPOINT " + savepoint)
		}
		if localtx != nil {
			localtx.Commit()
		}
		return nil
	}
	if err != nil {
		tx.Rollback()
		return newPropagationError(err, propagation)
	}
	if localtx != nil {
		localtx.Commit()
	}
	return nil
}
