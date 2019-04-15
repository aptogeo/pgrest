package pgrest_test

import (
	"context"
	"testing"
	"time"

	"github.com/aptogeo/pgrest"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type Todo struct {
	ID   uuid.UUID `sql:",pk"`
	Text string
}

func (t *Todo) BeforeInsert(c context.Context, db orm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

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
	Picture   []byte `sql:",type:bytea"`
	Books     []*Book
}

type PageOnly struct {
	NbPages int
}

func initTests(t *testing.T) (*pg.DB, *pgrest.Config) {
	db := pg.Connect(&pg.Options{
		User:               "postgres",
		Password:           "postgres",
		IdleCheckFrequency: 100 * time.Millisecond,
	})
	for _, model := range []interface{}{(*Author)(nil), (*Book)(nil), (*Todo)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			Temp: true,
		})
		assert.Nil(t, err)
	}

	config := pgrest.NewConfig("/rest/", db)
	config.AddResource(pgrest.NewResource("Todo", (*Todo)(nil), pgrest.All))
	config.AddResource(pgrest.NewResource("Author", (*Author)(nil), pgrest.All))
	config.AddResource(pgrest.NewResource("Book", (*Book)(nil), pgrest.All))
	return db, config
}

var todos = []Todo{{
	Text: "Todo1",
}, {
	Text: "Todo2",
}}

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
