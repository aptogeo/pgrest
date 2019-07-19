package pgrest

import (
	"context"
	"strings"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

// Executor structure
type Executor struct {
	restQuery          *RestQuery
	entity             interface{}
	count              int
	config             *Config
	originalSearchPath string
	ctx                context.Context
	tx                 *pg.Tx
}

// NewExecutor constructs Executor
func NewExecutor(config *Config, restQuery *RestQuery, entity interface{}) *Executor {
	e := new(Executor)
	e.restQuery = restQuery
	e.entity = entity
	e.count = 0
	e.config = config
	e.ctx = restQuery.Context()
	return e
}

func (e *Executor) begin() error {
	var err error
	e.tx, err = e.config.DB().WithContext(e.ctx).Begin()
	if err != nil {
		return err
	}
	return e.setSearchPath(e.restQuery.SearchPath)
}

func (e *Executor) commit() error {
	err := e.setSearchPath(e.originalSearchPath)
	if err != nil {
		return err
	}
	return e.tx.Commit()
}

func (e *Executor) rollback() error {
	return e.tx.Rollback()
}

func (e *Executor) getSearchPath() (string, error) {
	var searchPath string
	var err error
	_, err = e.tx.QueryOne(pg.Scan(&searchPath), "SHOW search_path")
	if err != nil {
		return "", err
	}
	searchPath = strings.Replace(searchPath, "\"\"", "\"", -1)
	searchPath = strings.Replace(searchPath, "\"\"", "\"", -1)
	return searchPath, nil
}

func (e *Executor) setSearchPath(searchPath string) error {
	var err error
	if searchPath == "" {
		return nil
	}
	if e.originalSearchPath == "" {
		e.originalSearchPath, err = e.getSearchPath()
		if err != nil {
			return err
		}
	}
	_, err = e.tx.Exec("SET search_path = " + searchPath)
	return err
}

func (e *Executor) getOne() error {
	q := e.tx.ModelContext(e.tx.Context(), e.entity).WherePK()
	q = addQueryFields(q, e.restQuery.Fields)
	if err := q.Select(); err != nil {
		return NewErrorFromCause(e.restQuery, err)
	}
	e.count = 1
	return nil
}

func (e *Executor) getSlice() error {
	var err error
	q := e.tx.ModelContext(e.tx.Context(), e.entity)
	q = addQueryLimit(q, e.restQuery.Limit)
	q = addQueryOffset(q, e.restQuery.Offset)
	q = addQueryFields(q, e.restQuery.Fields)
	q = addQuerySorts(q, e.restQuery.Sorts)
	q = addQueryFilter(q, e.restQuery.Filter, And)
	e.count, err = q.Count()
	if err != nil {
		return NewErrorFromCause(e.restQuery, err)
	}
	if e.count == 0 {
		return nil
	}
	if err = q.Select(); err != nil {
		return NewErrorFromCause(e.restQuery, err)
	}
	return nil
}

func (e *Executor) executeInsert() error {
	q := orm.NewQueryContext(e.tx.Context(), e.tx, e.entity)
	if _, err := q.Insert(); err != nil {
		return NewErrorFromCause(e.restQuery, err)
	}
	e.count = 1
	return nil
}

func (e *Executor) executeUpdate() error {
	q := orm.NewQueryContext(e.tx.Context(), e.tx, e.entity).WherePK()
	if _, err := q.Update(); err != nil {
		return NewErrorFromCause(e.restQuery, err)
	}
	e.count = 1
	return nil
}

func (e *Executor) executeDelete() error {
	q := orm.NewQueryContext(e.tx.Context(), e.tx, e.entity).WherePK()
	if _, err := q.Delete(e.entity); err != nil {
		return NewErrorFromCause(e.restQuery, err)
	}
	e.count = 1
	return nil
}
