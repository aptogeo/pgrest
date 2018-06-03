package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mathieumast/pgrest"
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
	r := gin.Default()
	r.Any("/rest/*r", func(c *gin.Context) {
		pgrestServer.ServeHTTP(c.Writer, c.Request)
	})
	err := r.Run()
	if err != nil {
		log.Fatal("Run: ", err)
	}
}
