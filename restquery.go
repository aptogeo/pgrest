package pgrest

import (
	"fmt"
	"strings"
)

// RestQuery structure
type RestQuery struct {
	Action      Action
	Resource    string
	Key         string
	ContentType string
	Content     []byte
	Offset      int
	Limit       int
	Fields      []*Field
	Sorts       []*Sort
	Filter      *Filter
}

func (q *RestQuery) String() string {
	if q.Action == Get {
		if q.Key == "" {
			return fmt.Sprintf("action=%v resource=%v offset=%v limit=%v fields=%v sorts=%v filter=%v", q.Action, q.Resource, q.Offset, q.Limit, q.Fields, q.Sorts, q.Filter)
		}
		return fmt.Sprintf("action=%v resource=%v key=%v fields=%v", q.Action, q.Resource, q.Key, q.Fields)
	}
	return fmt.Sprintf("action=%v resource=%v key=%v", q.Action, q.Resource, q.Key)
}

// Field structure
type Field struct {
	Name string
}

func (f *Field) String() string {
	return f.Name
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
	Op      Op        // operation
	Attr    string    // attribute name
	Value   string    // attribute value
	Filters []*Filter // sub filters for 'and', 'or' and 'not' operation
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
