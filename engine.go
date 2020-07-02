package pgrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/aptogeo/pgrest/transactional"
	"github.com/go-pg/pg/v9/orm"
	"github.com/vmihailenco/msgpack/v4"
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
		e.Config().InfoLogger().Printf("Execution request %v\n", restQuery)
	}
	resource, err := e.getResource(restQuery)
	if err != nil {
		return nil, &Error{Cause: err}
	}
	var entity interface{}
	var elem reflect.Value
	if restQuery.Action == Get {
		if restQuery.Key != "" {
			elem = reflect.New(resource.ResourceType()).Elem()
			entity = elem.Addr().Interface()
			if err = setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
				return nil, NewErrorFromCause(restQuery, err)
			}
		} else {
			sliceType := reflect.MakeSlice(reflect.SliceOf(resource.ResourceType()), 0, 0).Type()
			entity = reflect.New(sliceType).Interface()
		}
	} else if restQuery.Action == Post {
		if restQuery.Key != "" {
			return nil, NewErrorBadRequest("action 'Post': key is forbidden")
		}
		elem = reflect.New(resource.ResourceType()).Elem()
		entity = elem.Addr().Interface()
		if err = e.Deserialize(restQuery, resource, entity); err != nil {
			return nil, NewErrorFromCause(restQuery, err)
		}
	} else if restQuery.Action == Put {
		if restQuery.Key == "" {
			return nil, NewErrorBadRequest("action 'Put': key is mandatory")
		}
		elem = reflect.New(resource.ResourceType()).Elem()
		entity = elem.Addr().Interface()
		if err = e.Deserialize(restQuery, resource, entity); err != nil {
			return nil, NewErrorFromCause(restQuery, err)
		}
		setPk(resource.ResourceType(), elem, restQuery.Key)
	} else if restQuery.Action == Patch {
		if restQuery.Key == "" {
			return nil, NewErrorBadRequest("action 'Patch': key is mandatory")
		}
		elem = reflect.New(resource.ResourceType()).Elem()
		entity = elem.Addr().Interface()
		if err = setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
			return nil, NewErrorFromCause(restQuery, err)
		}
	} else if restQuery.Action == Delete {
		if restQuery.Key == "" {
			return nil, NewErrorBadRequest("action 'Delete': key is mandatory")
		}
		elem = reflect.New(resource.ResourceType()).Elem()
		entity = elem.Addr().Interface()
		if err = setPk(resource.ResourceType(), elem, restQuery.Key); err != nil {
			return nil, NewErrorFromCause(restQuery, err)
		}
	} else {
		return nil, &Error{Message: fmt.Sprintf("unknow action '%v'", restQuery.Action)}
	}

	var ctx context.Context
	if restQuery.Request != nil {
		ctx = restQuery.Request.Context()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = transactional.ContextWithDb(ctx, e.Config().DB())

	executor := NewExecutor(restQuery, entity)

	if restQuery.Action == Get {
		if restQuery.Key != "" {
			err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.GetOneExecFunc())
		} else {
			err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.GetSliceExecFunc())
		}
	} else if restQuery.Action == Post {
		err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.InsertExecFunc())
	} else if restQuery.Action == Put {
		err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.UpdateExecFunc())
	} else if restQuery.Action == Patch {
		err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.GetOneExecFunc())
		if err == nil {
			err = e.Deserialize(restQuery, resource, entity)
		}
		if err == nil {
			err = setPk(resource.ResourceType(), elem, restQuery.Key)
		}
		if err == nil {
			err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.UpdateExecFunc())
		}
	} else if restQuery.Action == Delete {
		err = executor.ExecuteWithSearchPath(ctx, restQuery.SearchPath, executor.DeleteExecFunc())
	}
	if err != nil {
		return nil, err
	}
	if restQuery.Debug {
		e.Config().InfoLogger().Printf("Execution result %v\n", entity)
	}
	if restQuery.Action == Get && restQuery.Key == "" {
		return NewPage(executor.entity, executor.count, restQuery), nil
	}
	return executor.entity, nil
}

// Deserialize deserializes data into entity
func (e *Engine) Deserialize(restQuery *RestQuery, resource *Resource, entity interface{}) error {
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
						field.ScanValue(elem, NewBytesReader([]byte(parts[1])), len(parts[1]))
						found = true
					}
				}
				if !found {
					for _, field := range table.Fields {
						if field.SQLName == parts[0] {
							field.ScanValue(elem, NewBytesReader([]byte(parts[1])), len(parts[1]))
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
		e.Config().ErrorLogger().Printf("Resource '%v' not defined in engine configuration", restQuery.Resource)
		e.Config().ErrorLogger().Printf("Configuration: '%v'", e.config)
		return nil, NewErrorBadRequest(fmt.Sprintf("resource '%v' not defined in engine configuration", restQuery.Resource))
	}
	return resource, nil
}
