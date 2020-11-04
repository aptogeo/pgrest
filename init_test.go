package pgrest_test

import (
	"context"
	"testing"

	"github.com/aptogeo/pgrest"
	"github.com/aptogeo/pgrest/transactional"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type Todo struct {
	ID   uuid.UUID `pg:",pk"`
	Text string
}

func (t *Todo) BeforeInsert(c context.Context) (context.Context, error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return c, nil
}

type Book struct {
	ID       int
	Title    string
	NbPages  int
	AuthorID int
	Author   *Author `pg:"rel:has-one"`
}

type Author struct {
	ID             int
	Firstname      string
	Lastname       string
	Picture        []byte  `pg:",type:bytea"`
	Books          []*Book `pg:"rel:has-many"`
	TransientField string  `pg:"-"`
}

func (b *Author) AfterSelect(ctx context.Context) error {
	tx := transactional.TxFromContext(ctx)
	var searchPathDB string
	tx.QueryOne(pg.Scan(&searchPathDB), "SHOW search_path")
	b.TransientField = searchPathDB
	return nil
}

type PageOnly struct {
	NbPages int
}

func initTests(t *testing.T) (*pg.DB, *pgrest.Config) {
	db := pg.Connect(&pg.Options{
		User: "postgres",
	})
	for _, model := range []interface{}{(*Author)(nil), (*Book)(nil), (*Todo)(nil)} {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
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
