package pgrest

import (
	"context"
	"strings"

	"github.com/aptogeo/pgrest/transactional"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// Executor structure
type Executor struct {
	restQuery          *RestQuery
	entity             interface{}
	count              int
	originalSearchPath string
}

// NewExecutor constructs Executor
func NewExecutor(restQuery *RestQuery, entity interface{}) *Executor {
	e := new(Executor)
	e.restQuery = restQuery
	e.entity = entity
	e.count = 0
	return e
}

func (e *Executor) GetSearchPath(ctx context.Context) (string, error) {
	var searchPath string
	var err error
	ctx, err = transactional.Execute(ctx, func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		tx.QueryOneContext(ctx, pg.Scan(&searchPath), "SHOW search_path")
		return ctx, nil
	})
	if err != nil {
		return "", err
	}
	searchPath = strings.Replace(searchPath, "\"\"", "\"", -1)
	searchPath = strings.Replace(searchPath, "\"\"", "\"", -1)
	return searchPath, nil
}

func (e *Executor) ExecuteWithSearchPath(ctx context.Context, searchPath string, execFunc transactional.ExecFunc) error {
	var err error
	if e.originalSearchPath == "" {
		e.originalSearchPath, err = e.GetSearchPath(ctx)
		if err != nil {
			return err
		}
	}
	ctx, err = transactional.Execute(ctx, func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		if searchPath != "" {
			_, err = tx.ExecContext(ctx, "SET search_path = "+searchPath)
			if err != nil {
				return ctx, err
			}
		}
		if execFunc != nil {
			ctx, err = execFunc(ctx, tx)
		}
		if searchPath != "" {
			tx.ExecContext(ctx, "SET search_path = "+e.originalSearchPath)
		}
		return ctx, err
	})
	return err
}

func (e *Executor) GetOneExecFunc() transactional.ExecFunc {
	return func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		q := tx.ModelContext(ctx, e.entity).WherePK()
		q = addQueryFields(q, e.restQuery.Fields)
		q = addQueryRelations(q, e.restQuery.Relations)
		if err := q.Select(); err != nil {
			return ctx, NewErrorFromCause(e.restQuery, err)
		}
		e.count = 1
		return ctx, nil
	}
}

func (e *Executor) GetSliceExecFunc() transactional.ExecFunc {
	return func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		var err error
		q := tx.ModelContext(ctx, e.entity)
		q = addQueryLimit(q, e.restQuery.Limit)
		q = addQueryOffset(q, e.restQuery.Offset)
		q = addQueryFields(q, e.restQuery.Fields)
		q = addQuerySorts(q, e.restQuery.Sorts)
		q = addQueryFilter(q, e.restQuery.Filter, And)
		e.count, err = q.Count()
		if err != nil {
			return ctx, NewErrorFromCause(e.restQuery, err)
		}
		if e.count == 0 {
			return ctx, nil
		}
		if err = q.Select(); err != nil {
			return ctx, NewErrorFromCause(e.restQuery, err)
		}
		return ctx, nil
	}
}

func (e *Executor) InsertExecFunc() transactional.ExecFunc {
	return func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		q := orm.NewQueryContext(ctx, tx, e.entity)
		if _, err := q.Insert(); err != nil {
			return ctx, NewErrorFromCause(e.restQuery, err)
		}
		e.count = 1
		return ctx, nil
	}
}

func (e *Executor) UpdateExecFunc() transactional.ExecFunc {
	return func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		q := orm.NewQueryContext(ctx, tx, e.entity).WherePK()
		if _, err := q.Update(); err != nil {
			return ctx, NewErrorFromCause(e.restQuery, err)
		}
		e.count = 1
		return ctx, nil
	}
}

func (e *Executor) DeleteExecFunc() transactional.ExecFunc {
	return func(ctx context.Context, tx *pg.Tx) (context.Context, error) {
		q := orm.NewQueryContext(ctx, tx, e.entity).WherePK()
		if _, err := q.Delete(e.entity); err != nil {
			return ctx, NewErrorFromCause(e.restQuery, err)
		}
		e.count = 1
		return ctx, nil
	}
}
