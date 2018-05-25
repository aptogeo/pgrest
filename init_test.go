package pgrest

import (
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/stretchr/testify/assert"
)

type Book struct {
	ID       int
	Title    string
	NbPages  int
	AuthorID int
	Author   *Author
}

type Author struct {
	ID        int
	Firstname string
	Lastname  string
	Books     []*Book
}

func initTests(t *testing.T) (*pg.DB, *Config) {
	db := pg.Connect(&pg.Options{
		User:               "postgres",
		Database:           "postgres",
		IdleCheckFrequency: 100 * time.Millisecond,
	})
	for _, model := range []interface{}{(*Author)(nil), (*Book)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			Temp: true,
		})
		assert.Nil(t, err)
	}

	config := NewConfig("/rest/", db)
	config.AddResource(NewResource("Author", (*Author)(nil), All))
	config.AddResource(NewResource("Book", (*Book)(nil), All))
	return db, config
}

var authors = []Author{{
	Firstname: "Antoine",
	Lastname:  "de Saint Exupéry",
}, {
	Firstname: "Franz",
	Lastname:  "Kafka",
}, {
	Firstname: "Francis Scott Key",
	Lastname:  "Fitzgerald",
}}

var books = []Book{{
	Title:    "Courrier sud",
	AuthorID: 1,
}, {
	Title:    "Vol de nuit",
	AuthorID: 1,
}, {
	Title:    "Terre des hommes",
	AuthorID: 1,
}, {
	Title:    "Lettre à un otage",
	AuthorID: 1,
}, {
	Title:    "Pilote de guerre",
	AuthorID: 1,
}, {
	Title:    "Le Petit Prince",
	AuthorID: 1,
}, {
	Title:    "La Métamorphose",
	AuthorID: 2,
}, {
	Title:    "La Colonie pénitentiaire",
	AuthorID: 2,
}, {
	Title:    "Le Procès",
	AuthorID: 2,
}, {
	Title:    "Le Château",
	AuthorID: 2,
}, {
	Title:    "L'Amérique",
	AuthorID: 2,
}, {
	Title:    "Gatsby le Magnifique",
	AuthorID: 3,
}}
