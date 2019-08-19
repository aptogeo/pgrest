package pgrest

import (
	"context"
	"fmt"
	"strings"
)

// RestQuery structure
type RestQuery struct {
	Action      Action
	Resource    string
	Key         string
	ContentType string
	Accept      string
	Content     []byte
	Offset      int
	Limit       int
	Fields      []*Field
	Relations   []*Relation
	Sorts       []*Sort
	Filter      *Filter
	SearchPath  string
	Debug       bool
	ctx         context.Context
}

// Context returns the request's context. To change the context, use
// WithContext.
func (q *RestQuery) Context() context.Context {
	if q.ctx != nil {
		return q.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of q with its context changed
// to ctx. The provided ctx must be non-nil.
func (q *RestQuery) WithContext(ctx context.Context) *RestQuery {
	if ctx == nil {
		panic("nil context")
	}
	q2 := new(RestQuery)
	*q2 = *q
	q2.ctx = ctx

	return q2
}

func (q *RestQuery) String() string {
	var str string
	if q.Action == Get {
		if q.Key == "" {
			str = fmt.Sprintf("action=%v resource=%v offset=%v limit=%v fields=%v relations=%v sorts=%v filter=%v", q.Action, q.Resource, q.Offset, q.Limit, q.Fields, q.Relations, q.Sorts, q.Filter)
		} else {
			str = fmt.Sprintf("action=%v resource=%v key=%v fields=%v relations=%v ", q.Action, q.Resource, q.Key, q.Fields, q.Relations)
		}
	} else if q.Action == Delete {
		str = fmt.Sprintf("action=%v resource=%v key=%v", q.Action, q.Resource, q.Key)
	} else {
		str = fmt.Sprintf("action=%v resource=%v key=%v content-type=%v content=%v", q.Action, q.Resource, q.Key, q.ContentType, q.Content)
	}
	if q.SearchPath != "" {
		str += fmt.Sprintf(" search_path=%v", q.SearchPath)
	}
	return str
}

// Field structure
type Field struct {
	Name string
}

func (f *Field) String() string {
	return f.Name
}

// Relation structure
type Relation struct {
	Name string
}

func (r *Relation) String() string {
	return r.Name
}

// Sort structure
type Sort struct {
	Name string
	Asc  bool
}

func (s *Sort) String() string {
	if s.Asc {
		return fmt.Sprintf("asc(%v)", s.Name)
	}
	return fmt.Sprintf("desc(%v)", s.Name)
}

// Filter structure
type Filter struct {
	Op      Op          // operation
	Attr    string      // attribute name
	Value   interface{} // attribute value
	Filters []*Filter   // sub filters for 'and', 'or' and 'not' operation
}

func (f *Filter) String() string {
	if f.Op == And || f.Op == Or {
		var sb strings.Builder
		for _, filter := range f.Filters {
			sb.WriteRune(' ')
			sb.WriteString(filter.String())
			sb.WriteRune(' ')
		}
		return fmt.Sprintf("%v (%v)", f.Op, sb)
	}
	return fmt.Sprintf("%v %v %v", f.Attr, f.Op, f.Value)
}
