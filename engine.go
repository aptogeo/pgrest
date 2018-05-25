package pgrest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-pg/pg/orm"
)

// Engine structure
type Engine struct {
	config *Config
}

// NewEngine constructs Engine
func NewEngine(config *Config) *Engine {
	e := new(Engine)
	e.config = config
	return e
}

// Config execugets config
func (e *Engine) Config() *Config {
	return e.config
}

// Execute executes a rest query
func (e *Engine) Execute(restQuery *RestQuery) (interface{}, error) {
	if restQuery.Resource == "" {
		return nil, fmt.Errorf("resource is mandatory")
	}
	resource, err := e.getResource(restQuery)
	if err != nil {
		return nil, err
	}
	if restQuery.Action|resource.Action() == None {
		return nil, fmt.Errorf("action query %v forbidden: resource action is %v", restQuery.Action, resource.Action())
	}

	if restQuery.Action == Get {
		return e.executeActionGet(resource, restQuery)
	} else if restQuery.Action == Post {
		return e.executeActionPost(resource, restQuery)
	} else if restQuery.Action == Put {
		return e.executeActionPut(resource, restQuery)
	} else if restQuery.Action == Patch {
		return e.executeActionPatch(resource, restQuery)
	} else if restQuery.Action == Delete {
		return e.executeActionDelete(resource, restQuery)
	}
	return nil, fmt.Errorf("unknow action: %v", restQuery.Action)
}

// Deserialize deserializes data into entity
func (e *Engine) Deserialize(restQuery *RestQuery, entity interface{}) error {
	resource, err := e.getResource(restQuery)
	if err != nil {
		return err
	}
	if regexp.MustCompile("[+-/]json($|[+-])").MatchString(restQuery.ContentType) {
		if err := json.Unmarshal(restQuery.Content, entity); err != nil {
			return err
		}
	} else if regexp.MustCompile("[+-/]form($|[+-])").MatchString(restQuery.ContentType) {
		table := orm.GetTable(resource.ResourceType())
		keyValues := strings.Split(string(restQuery.Content), "&")
		elem := reflect.ValueOf(entity).Elem()
		for _, keyValue := range keyValues {
			parts := strings.Split(keyValue, "=")
			if parts != nil && len(parts) == 2 {
				for _, field := range table.Fields {
					if field.GoName == parts[0] {
						field.ScanValue(elem, []byte(parts[1]))
					}
				}
			}
		}
	} else {
		return fmt.Errorf("no know content type: '%v'", restQuery.ContentType)
	}
	return nil
}

func (e *Engine) getResource(restQuery *RestQuery) (*Resource, error) {
	if restQuery.Resource == "" {
		return nil, fmt.Errorf("resource is mandatory")
	}
	resource := e.config.GetResource(restQuery.Resource)
	if resource == nil {
		return nil, fmt.Errorf("resource %v not found", restQuery.Resource)
	}
	return resource, nil
}

func (e *Engine) executeActionGet(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key != "" {
		return e.getOne(resource, restQuery)
	}
	return e.getPage(resource, restQuery)
}

func (e *Engine) executeActionPost(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key != "" {
		return nil, fmt.Errorf("action Post: key is forbidden")
	}
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	e.Deserialize(restQuery, entity)
	return entity, e.config.DB().Insert(entity)
}

func (e *Engine) executeActionPut(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key == "" {
		return nil, fmt.Errorf("action Put: key is mandatory")
	}
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	e.Deserialize(restQuery, entity)
	e.setPk(resource.ResourceType(), elem, restQuery.Key)
	return entity, e.config.DB().Update(entity)
}

func (e *Engine) executeActionPatch(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key == "" {
		return nil, fmt.Errorf("action Patch: key is mandatory")
	}
	entity, err := e.getOne(resource, restQuery)
	if err != nil {
		return nil, err
	}
	e.Deserialize(restQuery, entity)
	elem := reflect.ValueOf(entity).Elem()
	if err := e.setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
		return nil, err
	}
	return entity, e.config.DB().Update(entity)
}

func (e *Engine) executeActionDelete(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key == "" {
		return nil, fmt.Errorf("action Delete: key is mandatory")
	}
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	if err := e.setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
		return nil, err
	}
	return nil, e.config.DB().Delete(entity)
}

func (e *Engine) getOne(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	if err := e.setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
		return nil, err
	}
	query := e.config.DB().Model(entity).WherePK()
	e.addQueryFields(query, restQuery.Fields)
	if err := query.Select(); err != nil {
		return nil, err
	}
	return entity, nil
}

func (e *Engine) getPage(resource *Resource, restQuery *RestQuery) (*Page, error) {
	sliceType := reflect.MakeSlice(reflect.SliceOf(resource.ResourceType()), 0, 0).Type()
	entities := reflect.New(sliceType).Interface()
	query := e.config.DB().Model(entities)
	e.addQueryFields(query, restQuery.Fields)
	e.addQuerySorts(query, restQuery.Sorts)
	count, err := query.Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return NewPage(nil, 0, restQuery), nil
	}
	if err := query.Select(); err != nil {
		return nil, err
	}
	return NewPage(entities, uint64(count), restQuery), nil
}

func (e *Engine) setPk(resourceType reflect.Type, elem reflect.Value, key string) error {
	table := orm.GetTable(resourceType)
	if len(table.PKs) == 1 {
		pk := table.PKs[0]
		return pk.ScanValue(elem, []byte(key))
	}
	return fmt.Errorf("only single pk is permitted (resourse %v)", resourceType)
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
