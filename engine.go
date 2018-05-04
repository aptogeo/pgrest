package pgrest

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

// Engine structure
type Engine struct {
	config *Config
}

// Execute executes a rest query
func (e *Engine) Execute(db *pg.DB, restQuery *RestQuery) (res orm.Result, err error) {
	return nil, nil
}

// NewEngine constructs Engine
func NewEngine(config *Config) *Engine {
	e := new(Engine)
	e.config = config
	return e
}
