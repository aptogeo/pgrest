package pgrest

import (
	"fmt"
	"reflect"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

// Engine structure
type Engine struct {
	config *Config
}

// Execute executes a rest query
func (e *Engine) Execute(db *pg.DB, restQuery *RestQuery) (res interface{}, err error) {
	if restQuery.Resource == "" {
		return nil, fmt.Errorf("resource undefined")
	}
	resource := e.config.GetResource(restQuery.Resource)
	if resource == nil {
		return nil, fmt.Errorf("resource %v not found", restQuery.Resource)
	}
	if restQuery.Action|resource.Action == None {
		return nil, fmt.Errorf("action query %v forbidden: resource action is %v", restQuery.Action, resource.Action)
	}

	if restQuery.Action == Get {
		return e.executeActionGet(db, resource, restQuery)
	}
	return nil, nil
}

func (e *Engine) executeActionGet(db *pg.DB, resource *Resource, restQuery *RestQuery) (res interface{}, err error) {
	if restQuery.Key != "" {
		element := reflect.New(resource.Type)
		table := orm.GetTable(resource.Type)
		if len(table.PKs) == 1 {
			pk := table.PKs[0]
			pk.ScanValue(element, []byte(restQuery.Key))
		} else {
			return nil, fmt.Errorf("only single pk is permitted for %v", resource.Type)
		}
		db.Select(element)
		return element, nil
	}
	elements := reflect.ArrayOf(0, resource.Type)
	db.Model(elements).Select()
	return elements, nil
}

// NewEngine constructs Engine
func NewEngine(config *Config) *Engine {
	e := new(Engine)
	e.config = config
	return e
}
