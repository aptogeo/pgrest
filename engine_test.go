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
	ID       int
	Title    string
	AuthorID int
}

type Author struct {
	ID        int
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
	authors := []Author{{
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

	books := []Book{{
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
}

func TestEngine(t *testing.T) {
	config := NewConfig()
	config.AddResource(NewResource(reflect.TypeOf(Author{}), All))
	config.AddResource(NewResource(reflect.TypeOf(Book{}), All))
	engine := NewEngine(config)

	db := pg.Connect(pgOptions())
	defer db.Close()

	populate(db, t)

	var err error
	var res interface{}
	var authors []Author

	res, err = engine.Execute(db, &RestQuery{Get, "Author", "1", "", 0, 0, nil, nil})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.(*Author).ID, 1)
	assert.Equal(t, res.(*Author).Firstname, "Antoine")
	assert.Equal(t, len(res.(*Author).Books), 0)

	res, err = engine.Execute(db, &RestQuery{Get, "Author", "1", "", 0, 0, []Field{Field{"*"}, Field{"Books"}}, nil})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.(*Author).ID, 1)
	assert.Equal(t, res.(*Author).Firstname, "Antoine")
	assert.Equal(t, len(res.(*Author).Books), 6)
	assert.Equal(t, res.(*Author).Books[0].Title, "Courrier sud")

	res, err = engine.Execute(db, &RestQuery{Get, "Author", "2", "", 0, 0, []Field{Field{"*"}, Field{"Books"}}, nil})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.(*Author).ID, 2)
	assert.Equal(t, res.(*Author).Firstname, "Franz")
	assert.Equal(t, len(res.(*Author).Books), 5)
	assert.Equal(t, res.(*Author).Books[0].Title, "La Métamorphose")

	res, err = engine.Execute(db, &RestQuery{Get, "Author", "", "", 0, 10, nil, nil})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	authors = *res.(*[]Author)
	assert.Equal(t, len(authors), 2)
	assert.Equal(t, authors[0].ID, 1)

	res, err = engine.Execute(db, &RestQuery{Get, "Author", "", "", 0, 10, nil, []Sort{Sort{"firstname", true}}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	authors = *res.(*[]Author)
	assert.Equal(t, len(authors), 2)
	assert.Equal(t, authors[0].Firstname, "Antoine")
	assert.Equal(t, authors[1].Firstname, "Franz")

	res, err = engine.Execute(db, &RestQuery{Get, "Author", "", "", 0, 10, nil, []Sort{Sort{"firstname", false}}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	authors = *res.(*[]Author)
	assert.Equal(t, len(authors), 2)
	assert.Equal(t, authors[0].Firstname, "Franz")
	assert.Equal(t, authors[1].Firstname, "Antoine")
}
