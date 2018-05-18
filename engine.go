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
func (e *Engine) Execute(db *pg.DB, restQuery *RestQuery) (interface{}, error) {
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

func (e *Engine) executeActionGet(db *pg.DB, resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key != "" {
		return e.getOne(db, resource, restQuery)
	}
	return e.getPage(db, resource, restQuery)
}

func (e *Engine) getOne(db *pg.DB, resource *Resource, restQuery *RestQuery) (interface{}, error) {
	elem := reflect.New(resource.Type).Elem()
	table := orm.GetTable(resource.Type)
	if len(table.PKs) == 1 {
		pk := table.PKs[0]
		pk.ScanValue(elem, []byte(restQuery.Key))
	} else {
		return nil, fmt.Errorf("only single pk is permitted for %v", resource.Type)
	}
	iface := elem.Addr().Interface()
	query := db.Model(iface).WherePK()
	e.addQueryFields(query, restQuery.Fields)
	if err := query.Select(); err != nil {
		return nil, err
	}
	return iface, nil
}

func (e *Engine) getPage(db *pg.DB, resource *Resource, restQuery *RestQuery) (*Page, error) {
	sliceType := reflect.MakeSlice(reflect.SliceOf(resource.Type), 0, 0).Type()
	iface := reflect.New(sliceType).Interface()
	query := db.Model(iface)
	e.addQueryFields(query, restQuery.Fields)
	e.addQuerySorts(query, restQuery.Sorts)
	if err := query.Select(); err != nil {
		return nil, err
	}
	page := &Page{Slice: iface}
	return page, nil
}

func (e *Engine) addQueryFields(query *orm.Query, fields []Field) {
	if len(fields) > 0 {
		for _, field := range fields {
			query.Column(field.Name)
		}
	}
}

func (e *Engine) addQuerySorts(query *orm.Query, sorts []Sort) {
	if len(sorts) > 0 {
		for _, sort := range sorts {
			var orderExpr string
			if sort.Asc {
				orderExpr = sort.Name + " ASC"
			} else {
				orderExpr = sort.Name + " DESC"
			}
			query.Order(orderExpr)
		}
	}
}

// NewEngine constructs Engine
func NewEngine(config *Config) *Engine {
	e := new(Engine)
	e.config = config
	return e
}
