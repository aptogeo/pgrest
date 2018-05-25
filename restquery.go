package pgrest

import (
	"fmt"
)

// RestQuery structure
type RestQuery struct {
	Action      Action
	Resource    string
	Key         string
	ContentType string
	Content     []byte
	Offset      uint64
	Limit       uint64
	Fields      []Field
	Sorts       []Sort
}

func (q *RestQuery) String() string {
	if q.Action == Get {
		if q.Key == "" {
			return fmt.Sprintf("action=%v resource=%v offset=%v limit=%v fields=%v sorts=%v", q.Action, q.Resource, q.Offset, q.Limit, q.Fields, q.Sorts)
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
	Name  string
	Op    string
	Value string
}

func (f *Filter) String() string {
	return fmt.Sprintf("%v[%v]:%v", f.Name, f.Op, f.Value)
}
