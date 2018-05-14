package pgrest

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/stretchr/testify/assert"
)

type Book struct {
	ID       int64
	Title    string
	AuthorID int64
}

type Author struct {
	ID        int64
	Firstname string
	Lastname  string
	Books     []*Book
}

func pgOptions() *pg.Options {
	return &pg.Options{
		User:               "postgres",
		Database:           "postgres",
		IdleCheckFrequency: 100 * time.Millisecond,
	}
}

func populate(db *pg.DB, t *testing.T) {
	for _, model := range []interface{}{(*Author)(nil), (*Book)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			Temp: true,
		})
		assert.Nil(t, err)
	}
	authors := []*Author{{
		Firstname: "Antoine",
		Lastname:  "de Saint Exupéry",
	}, {
		Firstname: "Franz",
		Lastname:  "Kafka",
	}}
	err := db.Insert(&authors)
	assert.Nil(t, err)
	countAuthors, err := db.Model(&Author{}).Count()
	assert.Nil(t, err)
	assert.Equal(t, countAuthors, 2)

	books := []*Book{{
		Title:    "Courrier sud",
		AuthorID: authors[0].ID,
	}, {
		Title:    "Vol de nuit",
		AuthorID: authors[0].ID,
	}, {
		Title:    "Terre des hommes",
		AuthorID: authors[0].ID,
	}, {
		Title:    "Lettre à un otage",
		AuthorID: authors[0].ID,
	}, {
		Title:    "Pilote de guerre",
		AuthorID: authors[0].ID,
	}, {
		Title:    "Le Petit Prince",
		AuthorID: authors[0].ID,
	}, {
		Title:    "La Métamorphose",
		AuthorID: authors[1].ID,
	}, {
		Title:    "La Colonie pénitentiaire",
		AuthorID: authors[1].ID,
	}, {
		Title:    "Le Procès",
		AuthorID: authors[1].ID,
	}, {
		Title:    "Le Château",
		AuthorID: authors[1].ID,
	}, {
		Title:    "L'Amérique",
		AuthorID: authors[1].ID,
	}}

	err = db.Insert(&books)
	assert.Nil(t, err)
	countBooks, err := db.Model(&Book{}).Count()
	assert.Nil(t, err)
	assert.Equal(t, countBooks, 11)

	var selectedAuthors []Author
	err = db.Model(&selectedAuthors).
		Relation("Books").
		Select()
	assert.Nil(t, err)
	assert.Equal(t, len(selectedAuthors), 2)
	assert.Equal(t, len(selectedAuthors[0].Books), 6)
	assert.Equal(t, len(selectedAuthors[1].Books), 5)
}

func TestEngine(t *testing.T) {
	config := NewConfig()
	config.AddResource(NewResource(reflect.TypeOf(Author{}), All))
	config.AddResource(NewResource(reflect.TypeOf(Book{}), All))
	engine := NewEngine(config)

	db := pg.Connect(pgOptions())
	defer db.Close()

	populate(db, t)

	engine.Execute(db, &RestQuery{Get, "", "0", "", 0, 0, nil, nil})
}
