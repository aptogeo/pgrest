package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aptogeo/pgrest"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// Book struct
type Book struct {
	ID       int
	Title    string
	NbPages  int
	AuthorID int
	Author   *Author
}

// Author struct
type Author struct {
	ID        int
	Firstname string
	Lastname  string
	Books     []*Book
}

func newConfig() *pgrest.Config {
	db := pg.Connect(&pg.Options{
		User:               "postgres",
		Database:           "postgres",
		IdleCheckFrequency: 100 * time.Millisecond,
	})
	for _, model := range []interface{}{(*Author)(nil), (*Book)(nil)} {
		if err := db.CreateTable(model, &orm.CreateTableOptions{Temp: true}); err != nil {
			log.Fatal("CreateTable", err)
		}
	}

	config := pgrest.NewConfig("/rest/", db)
	config.AddResource(pgrest.NewResource("Author", (*Author)(nil), pgrest.All))
	config.AddResource(pgrest.NewResource("Book", (*Book)(nil), pgrest.All))
	return config
}

func main() {
	config := newConfig()
	defer config.DB().Close()
	pgrestServer := pgrest.NewServer(config)
	http.Handle("/rest/", pgrestServer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
