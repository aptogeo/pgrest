package pgrest

import (
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

// Resource structure
type Resource struct {
	name         string
	resourceType reflect.Type
	action       Action
}

// Name returns name
func (r *Resource) Name() string {
	return r.name
}

// ResourceType returns resourceType
func (r *Resource) ResourceType() reflect.Type {
	return r.resourceType
}

// Action returns action
func (r *Resource) Action() Action {
	return r.action
}

// NewResource constructs Resource
func NewResource(name string, entity interface{}, action Action) *Resource {
	orm.RegisterTable(entity)
	r := new(Resource)
	r.name = name
	r.resourceType = reflect.TypeOf(entity)
	if r.resourceType.Kind() == reflect.Ptr {
		r.resourceType = r.resourceType.Elem()
	}
	r.action = action
	return r
}

// Config structure
type Config struct {
	prefix             string
	db                 *pg.DB
	resources          map[string]*Resource
	defaultContentType string
	defaultAccept      string
	infoLogger         *log.Logger
	errorLogger        *log.Logger
}

// AddResource adds resource
func (c *Config) AddResource(resource *Resource) {
	c.resources[resource.Name()] = resource
}

// GetResource gets resource
func (c *Config) GetResource(resourceName string) *Resource {
	return c.resources[resourceName]
}

// SetPrefix sets prefix
func (c *Config) SetPrefix(prefix string) {
	c.prefix = prefix
	if c.prefix == "" {
		c.prefix = "/"
	}
	if !strings.HasPrefix(c.prefix, "/") {
		c.prefix = "/" + c.prefix
	}
	if !strings.HasSuffix(c.prefix, "/") {
		c.prefix = c.prefix + "/"
	}
}

// Prefix gets prefix
func (c *Config) Prefix() string {
	return c.prefix
}

// SetDefaultContentType sets defaultContentType
func (c *Config) SetDefaultContentType(defaultContentType string) {
	c.prefix = defaultContentType
}

// DefaultContentType gets defaultContentType
func (c *Config) DefaultContentType() string {
	return c.defaultContentType
}

// DefaultAccept gets defaultAccept
func (c *Config) DefaultAccept() string {
	return c.defaultAccept
}

// DB gets db
func (c *Config) DB() *pg.DB {
	return c.db
}

// SetInfoLogger sets info logger
func (c *Config) SetInfoLogger(logger *log.Logger) {
	c.infoLogger = logger
}

// InfoLogger gets info logger
func (c *Config) InfoLogger() *log.Logger {
	return c.infoLogger
}

// SetErrorLogger sets error logger
func (c *Config) SetErrorLogger(logger *log.Logger) {
	c.errorLogger = logger
}

// ErrorLogger gets error logger
func (c *Config) ErrorLogger() *log.Logger {
	return c.errorLogger
}

// NewConfig constructs Config
func NewConfig(prefix string, db *pg.DB) *Config {
	c := new(Config)
	c.prefix = prefix
	c.db = db
	c.resources = make(map[string]*Resource)
	c.defaultContentType = "application/json"
	c.defaultAccept = "application/json"
	c.infoLogger = log.New(os.Stdout, " INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	c.errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return c
}
