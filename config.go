package pgrest

import "reflect"

// Resource structure
type Resource struct {
	Type   reflect.Type
	Action Action
}

// NewResource constructs Resource
func NewResource(t reflect.Type, a Action) *Resource {
	r := new(Resource)
	r.Type = t
	r.Action = a
	return r
}

// Config structure
type Config struct {
	resources map[string]*Resource
}

// AddResource adds resource
func (c *Config) AddResource(r *Resource) {
	c.resources[r.Type.Name()] = r
}

// GetResource gets resource
func (c *Config) GetResource(resourceName string) *Resource {
	return c.resources[resourceName]
}

// NewConfig constructs Config
func NewConfig() *Config {
	c := new(Config)
	c.resources = make(map[string]*Resource)
	return c
}
