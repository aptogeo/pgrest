package pgrest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/types"
	"github.com/vmihailenco/msgpack"
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

// Config gets config
func (e *Engine) Config() *Config {
	return e.config
}

// Execute executes a rest query
func (e *Engine) Execute(restQuery *RestQuery) (interface{}, error) {
	if restQuery.Debug {
		e.Config().InfoLogger().Printf("Execute request %v\n", restQuery)
	}
	if restQuery.Resource == "" {
		return nil, NewErrorBadRequest("resource is mandatory")
	}
	resource, err := e.getResource(restQuery)
	if err != nil {
		return nil, &Error{Cause: err}
	}
	if restQuery.Action|resource.Action() == None {
		return nil, &Error{Message: fmt.Sprintf("action query '%v' forbidden: resource action is '%v'", restQuery.Action, resource.Action()), Code: 403}
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
	return nil, &Error{Message: fmt.Sprintf("unknow action '%v'", restQuery.Action)}
}

// Deserialize deserializes data into entity
func (e *Engine) Deserialize(restQuery *RestQuery, entity interface{}) error {
	resource, err := e.getResource(restQuery)
	if err != nil {
		return &Error{Cause: err}
	}
	if regexp.MustCompile("[+-/]json($|[+-;])").MatchString(restQuery.ContentType) {
		if err := json.Unmarshal(restQuery.Content, entity); err != nil {
			return &Error{Cause: err}
		}
	} else if regexp.MustCompile("[+-/]form($|[+-;])").MatchString(restQuery.ContentType) {
		table := orm.GetTable(resource.ResourceType())
		keyValues := strings.Split(string(restQuery.Content), "&")
		elem := reflect.ValueOf(entity).Elem()
		for _, keyValue := range keyValues {
			parts := strings.Split(keyValue, "=")
			if parts != nil && len(parts) == 2 {
				found := false
				for _, field := range table.Fields {
					if field.GoName == parts[0] {
						field.ScanValue(elem, types.NewBytesReader([]byte(parts[1])), len(parts[1]))
						found = true
					}
				}
				if !found {
					for _, field := range table.Fields {
						if field.SQLName == parts[0] {
							field.ScanValue(elem, types.NewBytesReader([]byte(parts[1])), len(parts[1]))
							found = true
						}
					}
				}
			}
		}
	} else if regexp.MustCompile("[+-/](msgpack|messagepack)($|[+-])").MatchString(restQuery.ContentType) {
		decoder := msgpack.NewDecoder(bytes.NewReader(restQuery.Content))
		decoder.UseJSONTag(true)
		if err := decoder.Decode(entity); err != nil {
			return &Error{Cause: err}
		}
	} else {
		return NewErrorBadRequest(fmt.Sprintf("Unknown content type '%v'", restQuery.ContentType))
	}
	if restQuery.Debug {
		e.Config().InfoLogger().Printf("Serialized response in %v: %v\n", restQuery.ContentType, entity)
	}
	return nil
}

func (e *Engine) getResource(restQuery *RestQuery) (*Resource, error) {
	if restQuery.Resource == "" {
		return nil, NewErrorBadRequest("resource is mandatory")
	}
	resource := e.config.GetResource(restQuery.Resource)
	if resource == nil {
		return nil, NewErrorBadRequest(fmt.Sprintf("resource '%v' not defined in engine configuration", restQuery.Resource))
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
		return nil, NewErrorBadRequest("action 'Post': key is forbidden")
	}
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	if err := e.Deserialize(restQuery, entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	if err := e.config.DB().WithContext(restQuery.Context()).Insert(entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	return entity, nil
}

func (e *Engine) executeActionPut(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key == "" {
		return nil, NewErrorBadRequest("action 'Put': key is mandatory")
	}
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	if err := e.Deserialize(restQuery, entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	setPk(resource.ResourceType(), elem, restQuery.Key)
	if err := e.config.DB().WithContext(restQuery.Context()).Update(entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	return entity, nil
}

func (e *Engine) executeActionPatch(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key == "" {
		return nil, NewErrorBadRequest("action 'Patch': key is mandatory")
	}
	entity, err := e.getOne(resource, restQuery)
	if err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	if err := e.Deserialize(restQuery, entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	elem := reflect.ValueOf(entity).Elem()
	if err := setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	if err := e.config.DB().WithContext(restQuery.Context()).Update(entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	return entity, nil
}

func (e *Engine) executeActionDelete(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	if restQuery.Key == "" {
		return nil, NewErrorBadRequest("action 'Delete': key is mandatory")
	}
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	if err := setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	if err := e.config.DB().WithContext(restQuery.Context()).Delete(entity); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	return entity, nil
}

func (e *Engine) getOne(resource *Resource, restQuery *RestQuery) (interface{}, error) {
	elem := reflect.New(resource.ResourceType()).Elem()
	entity := elem.Addr().Interface()
	if err := setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	q := e.config.DB().WithContext(restQuery.Context()).Model(entity).WherePK()
	q = addQueryFields(q, restQuery.Fields)
	if err := q.Select(); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	return entity, nil
}

func (e *Engine) getPage(resource *Resource, restQuery *RestQuery) (*Page, error) {
	sliceType := reflect.MakeSlice(reflect.SliceOf(resource.ResourceType()), 0, 0).Type()
	entities := reflect.New(sliceType).Interface()
	q := e.config.DB().WithContext(restQuery.Context()).Model(entities)
	q = addQueryLimit(q, restQuery.Limit)
	q = addQueryOffset(q, restQuery.Offset)
	q = addQueryFields(q, restQuery.Fields)
	q = addQuerySorts(q, restQuery.Sorts)
	q = addQueryFilter(q, restQuery.Filter, And)
	count, err := q.Count()
	if err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	if count == 0 {
		return NewPage(nil, 0, restQuery), nil
	}
	if err := q.Select(); err != nil {
		return nil, NewErrorFromCause(restQuery, err)
	}
	return NewPage(entities, count, restQuery), nil
}
