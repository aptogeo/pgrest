package main

import (
	"fmt"
)

// RestQuery structure
type RestQuery struct {
	Resource string
	Key      string
	Offset   uint64
	Limit    uint64
	Fields   []*Field
	Sorts    []*Sort
}

func (q *RestQuery) String() string {
	if q.Key != "" {
		return fmt.Sprintf("%v[%v]", q.Resource, q.Key)
	}
	return fmt.Sprintf("%v offset=%v limit=%v fields=%v sorts=%v", q.Resource, q.Offset, q.Limit, q.Fields, q.Sorts)
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
