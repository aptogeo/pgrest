package pgrest

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-pg/pg"
)

type Author struct {
	ID        int64
	Firstname string
	Lastname  string
	Books     []*Book // has many relation
}

type Book struct {
	ID     int64
	Title  string
	Resume string
	Author Author // has one relation
}

func pgOptions() *pg.Options {
	return &pg.Options{
		User:               "postgres",
		Database:           "postgres",
		IdleCheckFrequency: 100 * time.Millisecond,
	}
}

func TestEngine(t *testing.T) {
	config := NewConfig()
	config.AddResource(NewResource(reflect.TypeOf(Author{}), All))
	config.AddResource(NewResource(reflect.TypeOf(Book{}), All))
	engine := NewEngine(config)

	db := pg.Connect(pgOptions())
	defer db.Close()

	engine.Execute(db, &RestQuery{Get, "User", "1", "", 0, 0, nil, nil})
}
